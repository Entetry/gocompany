package handlers

import (
	"context"
	"fmt"
	cache2 "github.com/Entetry/gocompany/internal/cache"
	"github.com/Entetry/gocompany/internal/consumer"
	"github.com/Entetry/gocompany/internal/event"
	"github.com/Entetry/gocompany/internal/middleware"
	"github.com/Entetry/gocompany/internal/producer"
	"github.com/Entetry/gocompany/internal/repository"
	"github.com/Entetry/gocompany/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/ory/dockertest"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"testing"
	"time"
)

var (
	dbPool         *pgxpool.Pool
	companyHandler *Company
	e              *echo.Echo
	port           = 42800
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("unix:///home/entetry/.docker/desktop/docker.sock")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	pgResoursce, err := pool.Run("postgres", "14.1-alpine", []string{"POSTGRES_PASSWORD=password123"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	var dbHostAndPort string

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pool.Retry(func() error {
		dbHostAndPort = pgResoursce.GetHostPort("5432/tcp")

		dbPool, err = pgxpool.Connect(ctx, fmt.Sprintf("postgresql://postgres:password123@%v/postgres", dbHostAndPort))
		if err != nil {
			return err
		}

		return dbPool.Ping(ctx)
	})
	if err != nil {
		cancel()
		log.Fatalf("Could not connect to database: %s", err)
	}
	cmd := exec.Command("flyway",
		"-user=postgres",
		"-password=password123",
		"-locations=filesystem:../../migrations",
		fmt.Sprintf("-url=jdbc:postgresql://%v/postgres", dbHostAndPort),
		"migrate")

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	var redisClient *redis.Client
	redisRsc, err := pool.Run("redis", "7-alpine", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		redisClient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", redisRsc.GetPort("6379/tcp")),
		})

		return redisClient.Ping(ctx).Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	companyRepository := repository.NewCompanyRepository(dbPool)
	logoRepository := repository.NewLogoRepository(dbPool)
	cacheCompany := cache2.NewLocalCache()
	redisProducer := producer.NewRedisCompanyProducer(redisClient)
	companyService := service.NewCompany(companyRepository, logoRepository, cacheCompany, redisProducer)
	companyHandler = NewCompany(companyService)
	go ConsumeCompanies(redisClient, cacheCompany)
	e = echo.New()
	e.Validator = middleware.NewCustomValidator(validator.New())
	company := e.Group("api/company")
	company.POST("", companyHandler.Create)
	company.GET("", companyHandler.GetAll)
	company.GET("/:id", companyHandler.GetByID)
	company.PUT("", companyHandler.Update)
	company.DELETE("/:id", companyHandler.Delete)
	company.POST("/logo", companyHandler.AddLogo)
	company.GET("/logo/:id", companyHandler.GetLogoByCompanyID)
	go func() {
		err = e.Start(fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatal(err)
		}
	}()
	defer func(e *echo.Echo, ctx context.Context) {
		err = e.Shutdown(ctx)
		if err != nil {
			log.Errorf("Error while stopping server \n %v", err)
		}
	}(e, ctx)
	code := m.Run()
	resources := []*dockertest.Resource{pgResoursce, redisRsc}
	for _, resource := range resources {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s\n", err)
		}
		err = resource.Expire(1)
		if err != nil {
			log.Print(err)
		}
	}

	os.Exit(code)
}

func ConsumeCompanies(redisClient *redis.Client, localCache *cache2.LocalCache) {
	redisCompanyConsumer := consumer.NewRedisCompanyConsumer(redisClient, fmt.Sprintf("%d000-0", time.Now().Unix()))
	go redisCompanyConsumer.Consume(context.Background(), func(id uuid.UUID, action, name string) {
		switch action {
		case event.UPDATE:
			localCache.Update(id, name)
		case event.DELETE:
			localCache.Delete(id)
		default:
			log.Error("Unknown event")
		}
	})
}
