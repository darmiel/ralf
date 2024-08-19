//go:build server
// +build server

package main

import (
	v1 "github.com/darmiel/ralf/cmd/server/api/v1"
	"github.com/darmiel/ralf/cmd/server/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
	"net/http"
	"os"
	"strconv"
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

	// Health check
	httpApp.Get("/icanhazralf", func(ctx *fiber.Ctx) error {
		return ctx.Status(http.StatusOK).JSON(buildInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		})
	})

	// Setup services
	storageService := setupStorageService()
	authService := setupAuthService()
	cacheService := setupCacheService()

	v1Group := httpApp.Group("/v1")
	v1.RegisterRoutes(v1Group, storageService, authService, cacheService)

	// the current newest version of the API is v1, so we can redirect the root to it
	v1.RegisterRoutes(httpApp, storageService, authService, cacheService)
}

// setupStorageService initializes the storage service based on the environment variables.
func setupStorageService() service.StorageService {
	storageProvider := os.Getenv("STORAGE_PROVIDER")
	if storageProvider == "" {
		return nil
	}

	switch storageProvider {
	case "mongo":
		log.Println("Using MongoDB storage service")
		mongoURI := os.Getenv("MONGODB_URI")
		dbName := os.Getenv("MONGODB_DB")
		if mongoURI == "" || dbName == "" {
			log.Fatal("MongoDB configuration is required for storage service")
		}
		return service.NewMongoStorageService(mongoURI, dbName)
	default:
		log.Fatalln("Unknown storage service:", storageProvider)
		return nil // No storage service is provided
	}
}

// setupAuthService initializes the authentication service based on the environment variables.
func setupAuthService() service.AuthService {
	authProvider := os.Getenv("AUTH_PROVIDER")
	if authProvider == "" {
		return nil
	}

	switch authProvider {
	case "firebase":
		log.Println("Using Firebase authentication service")
		firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS")
		if firebaseCredentials == "" {
			log.Fatal("Firebase credentials file is required for Firebase authentication")
		}
		authService, err := service.NewFirebaseAuthService(firebaseCredentials)
		if err != nil {
			log.Fatalf("Failed to initialize FirebaseAuthService: %v", err)
		}
		return authService
	default:
		log.Fatalln("Unknown auth service:", authProvider)
		return nil // No auth service is provided
	}
}

// setupCacheService initializes the cache service based on the environment variables.
func setupCacheService() service.CacheService {
	cacheProvider := os.Getenv("CACHE_PROVIDER")

	switch cacheProvider {
	case "redis":
		log.Println("Using Redis cache service")
		redisAddr := os.Getenv("REDIS_ADDR")
		redisPassword := os.Getenv("REDIS_PASSWORD")
		redisDBStr := os.Getenv("REDIS_DB")

		redisDB, err := strconv.Atoi(redisDBStr)
		if err != nil {
			log.Fatalf("Invalid REDIS_DB value: %v", err)
		}

		return service.NewRedisCacheService(redisAddr, redisPassword, redisDB)
	default:
		log.Println("Using in-memory cache service")
		// Use in-memory cache by default
		return service.NewLocalCacheService(5*time.Minute, 10*time.Minute)
	}
}
