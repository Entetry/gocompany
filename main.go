// Package main
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "entetry/gotest/docs"
	"entetry/gotest/internal/cache"
	"entetry/gotest/internal/config"
	"entetry/gotest/internal/consumer"
	"entetry/gotest/internal/event"
	"entetry/gotest/internal/handlers"
	"entetry/gotest/internal/middleware"
	"entetry/gotest/internal/producer"
	"entetry/gotest/internal/repository/postgre"
	"entetry/gotest/internal/service"
)

// @title          Gotest Swagger API
// @version        1.0
// @description    Swagger API for Golang Project gotest.
// @termsOfService http://swagger.io/terms/

// @contact.name  API Support
// @contact.email antonklintsevich@gmail.com

// @BasePath /api
func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	jwtCfg, err := config.NewJwtConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	db, err := pgxpool.Connect(ctx, cfg.ConnectionString)
	if err != nil {
		log.Fatalf("Couldn't connect to database: %s\n", err) //nolint:errcheck,gocritic
	}
	defer db.Close()

	redisClient := buildRedis(cfg)
	defer func(redisClient *redis.Client) {
		redisErr := redisClient.Close()
		if redisErr != nil {
			log.Error(err)
		}
	}(redisClient)

	refreshSessionRepository := postgre.NewRefresh(db)
	refreshSessionService := service.NewRefreshSession(refreshSessionRepository)

	userRepository := postgre.NewUserRepository(db)
	userService := service.NewUserService(userRepository)

	authService := service.NewAuthService(userService, refreshSessionService, jwtCfg)
	authHandler := handlers.NewAuth(authService)

	redisProducer := producer.NewRedisCompanyProducer(redisClient)
	cacheCompany := cache.NewLocalCache()

	companyRepository := postgre.NewCompanyRepository(db)
	logoRepository := postgre.NewLogoRepository(db)
	companyService := service.NewCompany(companyRepository, logoRepository, cacheCompany, redisProducer)
	companyHandler := handlers.NewCompany(companyService)

	go consumeCompanies(redisClient, cacheCompany)

	e := echo.New()

	e.Validator = middleware.NewCustomValidator(validator.New())
	auth := e.Group("api/auth")
	auth.POST("/refresh-tokens", authHandler.Refresh)
	auth.POST("/sign-in", authHandler.SignIn)
	auth.POST("/sign-up", authHandler.SignUp)
	auth.POST("/logout", authHandler.Logout)

	company := e.Group("api/company")
	company.Use(middleware.NewJwtMiddleware(jwtCfg.AccessTokenKey))
	company.POST("", companyHandler.Create)
	company.GET("", companyHandler.GetAll)
	company.GET("/:id", companyHandler.GetByID)
	company.PUT("", companyHandler.Update)
	company.DELETE("/:id", companyHandler.Delete)
	company.POST("/logo", companyHandler.AddLogo)
	company.GET("/logo/:id", companyHandler.GetLogoByCompanyID)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	err = e.Start(fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Server started on ", cfg.Port)
	go func() {
		<-sigChan
		cancel()
		err = e.Shutdown(ctx)
		if err != nil {
			log.Errorf("can't stop server gracefully %v", err)
		}
	}()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal(err)
	}
	err = e.Server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}

func consumeCompanies(redisClient *redis.Client, localCache *cache.LocalCache) {
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

func buildRedis(cfg *config.Config) *redis.Client {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPass,
	}

	redisClient := redis.NewClient(opts)
	_, err := redisClient.Ping(context.Background()).Result()

	if err != nil {
		log.Fatal(err)
	}

	return redisClient
}
