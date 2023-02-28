// Package main
package main

import (
	"context"
	"fmt"
	"github.com/Entetry/gocompany/internal/middleware"
	"github.com/Entetry/gocompany/internal/repository"
	"github.com/Entetry/gocompany/protocol/companyService"
	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Entetry/gocompany/docs"
	"github.com/Entetry/gocompany/internal/cache"
	"github.com/Entetry/gocompany/internal/config"
	"github.com/Entetry/gocompany/internal/consumer"
	"github.com/Entetry/gocompany/internal/event"
	"github.com/Entetry/gocompany/internal/handlers"
	"github.com/Entetry/gocompany/internal/producer"
	"github.com/Entetry/gocompany/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

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

	refreshSessionRepository := repository.NewRefresh(db)
	refreshSessionService := service.NewRefreshSession(refreshSessionRepository)

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)

	authService := service.NewAuthService(userService, refreshSessionService, jwtCfg)
	authHandler := handlers.NewAuth(authService)

	redisProducer := producer.NewRedisCompanyProducer(redisClient)
	cacheCompany := cache.NewLocalCache()

	companyRepository := repository.NewCompanyRepository(db)
	logoRepository := repository.NewLogoRepository(db)
	cmpService := service.NewCompany(companyRepository, logoRepository, cacheCompany, redisProducer)
	cmpHandler := handlers.NewCompany(cmpService)

	go ConsumeCompanies(redisClient, cacheCompany)
	jwtInterceptor := middleware.NewAuthInterceptor(jwtCfg)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(jwtInterceptor.Unary), grpc.StreamInterceptor(jwtInterceptor.StreamInterceptor))
	companyService.RegisterCompanyServiceServer(grpcServer, cmpHandler)
	companyService.RegisterAuthGRPCServiceServer(grpcServer, authHandler)
	go func() {
		<-sigChan
		cancel()
		grpcServer.GracefulStop()
		if err != nil {
			log.Errorf("can't stop server gracefully %v", err)
		}
	}()
	log.Info("grpc Server started on ", cfg.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal(err)
	}
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func ConsumeCompanies(redisClient *redis.Client, localCache *cache.LocalCache) {
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
