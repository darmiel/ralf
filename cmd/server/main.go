//go:build server
// +build server

package main

import (
	"fmt"
	"github.com/darmiel/ralf/cmd/server/api/middlewares"
	v1 "github.com/darmiel/ralf/cmd/server/api/v1"
	"github.com/darmiel/ralf/cmd/server/logging"
	"github.com/darmiel/ralf/cmd/server/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	version string
	commit  string
	date    string
)

type buildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func main() {
	logger := logging.Get()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	httpApp := fiber.New(fiber.Config{
		AppName:                 "E. T.",
		EnableTrustedProxyCheck: false,
		ProxyHeader:             "X-Forwarded-For",
	})

	httpApp.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}))

	httpApp.Use(middlewares.LogMiddleware)

	// Health check
	httpApp.Get("/icanhazralf", func(ctx *fiber.Ctx) error {
		return ctx.Status(http.StatusOK).JSON(buildInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		})
	})

	// Setup services
	storageService := setupStorageService(logger)
	authService := setupAuthService(logger)
	cacheService := setupCacheService(logger)

	v1Group := httpApp.Group("/v1")
	v1.RegisterRoutes(v1Group, storageService, authService, cacheService)

	// the current newest version of the API is v1, so we can redirect the root to it
	v1.RegisterRoutes(httpApp, storageService, authService, cacheService)

	// start the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-c
		logger.Info("Gracefully shutting down...")
		if err := httpApp.Shutdown(); err != nil {
			logger.Fatal("Cannot shutdown server", zap.Error(err))
		}
		logger.Info("Bye!")
	}()

	var bind string
	if logging.IsProduction() {
		bind = ":8080"
	} else {
		bind = "127.0.0.1:8080"
	}

	httpApp.Hooks().OnListen(func(data fiber.ListenData) error {
		logger.Info(fmt.Sprintf("Server is running on %s:%s", data.Host, data.Port))
		return nil
	})

	logger.Info(fmt.Sprintf("Starting server on %s", bind))
	if err := httpApp.Listen(bind); err != nil {
		logger.Fatal("Cannot bind address", zap.Error(err))
	}
}

// setupStorageService initializes the storage service based on the environment variables.
func setupStorageService(logger *zap.Logger) service.StorageService {
	storageProvider := os.Getenv("STORAGE_PROVIDER")
	if storageProvider == "" {
		return nil
	}

	switch storageProvider {
	case "mongo":
		logger.Info("Using MongoDB storage service")
		mongoURI := os.Getenv("MONGODB_URI")
		dbName := os.Getenv("MONGODB_DB")
		if mongoURI == "" || dbName == "" {
			logger.Fatal("MongoDB configuration is required for storage service")
		}
		return service.NewMongoStorageService(mongoURI, dbName)
	default:
		logger.Fatal("Unknown storage service", zap.String("service", storageProvider))
		return nil // No storage service is provided
	}
}

// setupAuthService initializes the authentication service based on the environment variables.
func setupAuthService(logger *zap.Logger) service.AuthService {
	authProvider := os.Getenv("AUTH_PROVIDER")
	if authProvider == "" {
		return nil
	}

	switch authProvider {
	case "firebase":
		logger.Info("Using Firebase authentication service")
		firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS_PATH")
		if firebaseCredentials == "" {
			logger.Fatal("Firebase credentials file is required for Firebase authentication")
		}
		authService, err := service.NewFirebaseAuthService(firebaseCredentials)
		if err != nil {
			logger.Fatal("Failed to initialize FirebaseAuthService", zap.Error(err))
		}
		return authService
	default:
		logger.Fatal("Unknown auth service", zap.String("service", authProvider))
		return nil // No auth service is provided
	}
}

// setupCacheService initializes the cache service based on the environment variables.
func setupCacheService(logger *zap.Logger) service.CacheService {
	cacheProvider := os.Getenv("CACHE_PROVIDER")

	switch cacheProvider {
	case "redis":
		logger.Info("Using Redis cache service")
		redisAddr := os.Getenv("REDIS_ADDR")
		redisPassword := os.Getenv("REDIS_PASSWORD")
		redisDBStr := os.Getenv("REDIS_DB")

		redisDB, err := strconv.Atoi(redisDBStr)
		if err != nil {
			logger.Fatal("Invalid REDIS_DB value", zap.Error(err))
		}

		return service.NewRedisCacheService(redisAddr, redisPassword, redisDB)
	default:
		logger.Info("Using in-memory cache service")
		// Use in-memory cache by default
		return service.NewLocalCacheService(5*time.Minute, 10*time.Minute)
	}
}
