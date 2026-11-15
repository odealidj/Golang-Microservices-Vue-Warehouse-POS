package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	jwtConf "micro-warehouse/api-gateway/config"
	"micro-warehouse/api-gateway/controller"
	"micro-warehouse/api-gateway/middleware"
	"micro-warehouse/observability"
)

var proxyHTTPClient = observability.NewTracedHTTPClient(30 * time.Second)

type ServiceConfig struct {
	Name string
	URL  string
}

type Config struct {
	Services map[string]ServiceConfig
	Port     string
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found, trying env.example")
		if err := godotenv.Load("env.example"); err != nil {
			log.Println("Warning: env.example file not found, using default configuration")
		}
	}

	config := loadConfig()
	jwtConfig := jwtConf.LoadJWTConfig()
	redisConfig := jwtConf.LoadRedisConfig()

	shutdownTracing, err := observability.InitTracing(context.Background(), "api-gateway")
	if err != nil {
		log.Printf("Warning: failed to initialize tracing: %v", err)
	} else {
		defer func() {
			shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTracing(shutCtx)
		}()
	}

	var redisClient *redis.Client
	if redisConfig.Host != "" {
		candidate := jwtConf.NewRedisClient(redisConfig)
		if err := jwtConf.TestRedisConnection(candidate); err != nil {
			log.Printf("Warning: Redis not reachable (%v), falling back to in-memory rate limiter", err)
			_ = candidate.Close()
		} else {
			redisClient = candidate
		}
	}

	app := buildApp(config, jwtConfig, redisClient)

	log.Printf("🚀 API Gateway starting on port %s", config.Port)
	log.Fatal(app.Listen(":" + config.Port))
}

func buildApp(config Config, jwtConfig middleware.JWTConfig, redisClient *redis.Client) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Warehouse Project API Gateway",
		ServerHeader: "Warehouse-API-Gateway",
	})

	memRateConfig := middleware.RateLimiterConfig{
		GlobalMax:        1000,
		GlobalExpiration: 1 * time.Minute,
		AuthMax:          10,
		AuthExpiration:   1 * time.Minute,
		APIMax:           100,
		APIExpiration:    1 * time.Minute,
	}

	redisRateConfig := middleware.RedisRateLimiterConfig{
		Max:         memRateConfig.APIMax,
		Expiration:  memRateConfig.APIExpiration,
		RedisClient: redisClient,
	}

	if redisClient != nil {
		globalCfg := redisRateConfig
		globalCfg.Max = memRateConfig.GlobalMax
		app.Use(middleware.RedisGlobalRateLimiter(globalCfg))
	} else {
		app.Use(middleware.GlobalRateLimiter(memRateConfig))
	}
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(observability.TracingMiddleware("api-gateway"))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"message": "API Gateway is running",
			"services": fiber.Map{
				"user-service":         config.Services["user"].URL,
				"product-service":      config.Services["product"].URL,
				"merchant-service":     config.Services["merchant"].URL,
				"notification-service": config.Services["notification"].URL,
				"transaction-service":  config.Services["transaction"].URL,
				"midtrans-service":     config.Services["midtrans"].URL,
				"warehouse-service":    config.Services["warehouse"].URL,
				"role-service":         config.Services["role"].URL,
				"assign-role-service":  config.Services["assign-role"].URL,
			},
		})
	})

	authController := controller.NewAuthController(config.Services["auth"].URL, jwtConfig)

	setUpAuthRoutes(app, authController, redisRateConfig, memRateConfig)
	setupMidtransCallbackRoute(app, config.Services["midtrans"])

	setupProtectedRoutes(app, config, jwtConfig, redisRateConfig, memRateConfig)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"error":   "Not Found",
			"message": "Service not found",
			"path":    c.Path(),
		})
	})

	return app
}

func loadConfig() Config {
	port := getEnv("PORT", "")
	if port == "" {
		port = getEnv("APP_PORT", "8080")
	}

	config := Config{
		Port: port,
		Services: map[string]ServiceConfig{
			"user": {
				Name: "user-service",
				URL:  getEnv("USER_SERVICE_URL", "http://localhost:8081"),
			},
			"role": {
				Name: "role-service",
				URL:  getEnv("USER_SERVICE_URL", "http://localhost:8081"),
			},
			"assign-role": {
				Name: "assign-role-service",
				URL:  getEnv("USER_SERVICE_URL", "http://localhost:8081"),
			},
			"auth": {
				Name: "auth-service",
				URL:  getEnv("USER_SERVICE_URL", "http://localhost:8081"),
			},
			"product": {
				Name: "product-service",
				URL:  getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
			},
			"merchant": {
				Name: "merchant-service",
				URL:  getEnv("MERCHANT_SERVICE_URL", "http://localhost:8084"),
			},
			"notification": {
				Name: "notification-service",
				URL:  getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8086"),
			},
			"transaction": {
				Name: "transaction-service",
				URL:  getEnv("TRANSACTION_SERVICE_URL", "http://localhost:8085"),
			},
			"midtrans": {
				Name: "midtrans-service",
				URL:  getEnv("TRANSACTION_SERVICE_URL", "http://localhost:8085"),
			},
			"warehouse": {
				Name: "warehouse-service",
				URL:  getEnv("WAREHOUSE_SERVICE_URL", "http://localhost:8083"),
			},
		},
	}

	log.Println("📋 Service Configuration:")
	for name, service := range config.Services {
		log.Printf("  %s: %s", name, service.URL)
	}

	return config
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func setUpAuthRoutes(app *fiber.App, authController *controller.AuthController, rateLimiterConfig middleware.RedisRateLimiterConfig, memRateConfig middleware.RateLimiterConfig) {
	authGroup := app.Group("/api/v1/auth")

	if rateLimiterConfig.RedisClient != nil {
		authCfg := rateLimiterConfig
		authCfg.Max = memRateConfig.AuthMax
		authGroup.Use(middleware.RedisAuthRateLimiter(authCfg))
	} else {
		authGroup.Use(middleware.AuthRateLimiter(memRateConfig))
	}
	authGroup.Post("/login", authController.Login)
}

func setupMidtransCallbackRoute(app *fiber.App, service ServiceConfig) {
	app.Post("/api/v1/midtrans/callback", func(c *fiber.Ctx) error {
		fullURL := service.URL + "/api/v1/midtrans/callback"
		return forwardProxyRequest(c, service.URL, fullURL, false)
	})
}

func proxyRequestWithPath(c *fiber.Ctx, targetURL string, basePath string) error {
	fullURL := targetURL + c.Path()
	if queryParams := c.Context().QueryArgs().String(); queryParams != "" {
		fullURL += "?" + queryParams
	}
	return forwardProxyRequest(c, targetURL, fullURL, true)
}

func proxyRequest(c *fiber.Ctx, targetURL string) error {
	path := c.Params("*")
	if path == "" {
		path = c.Path()
	}

	path = strings.TrimPrefix(path, "api/v1/")

	fullURL := targetURL
	if !strings.HasSuffix(targetURL, "/") && !strings.HasPrefix(path, "/") {
		fullURL += "/"
	}
	fullURL += path

	if queryParams := c.Context().QueryArgs().String(); queryParams != "" {
		fullURL += "?" + queryParams
	}

	return forwardProxyRequest(c, targetURL, fullURL, true)
}

func forwardProxyRequest(c *fiber.Ctx, targetURL, fullURL string, includeUserHeaders bool) error {
	req, err := http.NewRequestWithContext(c.UserContext(), c.Method(), fullURL, bytes.NewReader(c.Body()))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to create request",
		})
	}

	for key, values := range c.GetReqHeaders() {
		// Strip caller-supplied internal/identity headers to prevent header spoofing.
		switch strings.ToLower(key) {
		case "x-gateway", "x-internal-request", "x-user-id", "x-user-email", "x-user-roles":
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	req.Header.Set("X-Gateway", "warehouse-api-gateway")
	req.Header.Set("X-Internal-Request", "true")

	if includeUserHeaders {
		addUserHeaders(req, c)
	}

	resp, err := proxyHTTPClient.Do(req)
	if err != nil {
		log.Printf("Error making request to %s: %v", fullURL, err)
		return c.Status(502).JSON(fiber.Map{
			"error":   "Bad Gateway",
			"message": "Service unavailable",
			"service": targetURL,
		})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error":   "Internal Server Error",
			"message": "Failed to read response",
		})
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	return c.Status(resp.StatusCode).Send(respBody)
}

func addUserHeaders(req *http.Request, c *fiber.Ctx) {
	if userID := c.Locals("user_id"); userID != nil {
		req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
	}
	if userEmail := c.Locals("user_email"); userEmail != nil {
		req.Header.Set("X-User-Email", fmt.Sprintf("%v", userEmail))
	}
	if userRoles := c.Locals("user_roles"); userRoles != nil {
		req.Header.Set("X-User-Roles", fmt.Sprintf("%v", userRoles))
	}
}

func setupProtectedRoutes(app *fiber.App, config Config, jwtConfig middleware.JWTConfig, rateLimiterConfig middleware.RedisRateLimiterConfig, memRateConfig middleware.RateLimiterConfig) {
	protected := app.Group("/api/v1", middleware.JWTAuthMiddleware(jwtConfig))

	if rateLimiterConfig.RedisClient != nil {
		protected.Use(middleware.RedisAPIRateLimiter(rateLimiterConfig))
	} else {
		protected.Use(middleware.APIRateLimiter(memRateConfig))
	}

	setupUserRoutes(protected, config.Services["user"])
	setupRoleRoutes(protected, config.Services["role"])
	setupAssignRoleRoutes(protected, config.Services["assign-role"])
	setupProductRoutes(protected, config.Services["product"])
	setupMerchantRoutes(protected, config.Services["merchant"])
	setupTransactionRoutes(protected, config.Services["transaction"])
	setupWarehouseRoutes(protected, config.Services["warehouse"])
}

func setupUserRoutes(router fiber.Router, service ServiceConfig) {
	userGroup := router.Group("/users")

	userGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/users")
	})

	userGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/users")
	})

	uploadUserGroup := router.Group("/upload")
	uploadUserGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/upload")
	})

	uploadUserGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/upload")
	})
}

func setupRoleRoutes(router fiber.Router, service ServiceConfig) {
	roleGroup := router.Group("/roles")

	roleGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/roles")
	})

	roleGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/roles")
	})
}

func setupAssignRoleRoutes(router fiber.Router, service ServiceConfig) {
	assignRoleGroup := router.Group("/assign-role")

	assignRoleGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/assign-role")
	})

	assignRoleGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/assign-role")
	})
}

func setupProductRoutes(router fiber.Router, service ServiceConfig) {
	productGroup := router.Group("/products")

	productGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/products")
	})

	productGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/products")
	})

	categoryGroup := router.Group("/categories")

	categoryGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/categories")
	})

	categoryGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/categories")
	})

	uploadProductGroup := router.Group("/upload-product")
	uploadProductGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/upload-product")
	})

	uploadProductGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/upload-product")
	})
}

func setupMerchantRoutes(router fiber.Router, service ServiceConfig) {
	merchantGroup := router.Group("/merchants")

	merchantGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/merchants")
	})

	merchantGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/merchants")
	})

	merchantProductGroup := router.Group("/merchant-products")
	merchantProductGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/merchant-products")
	})

	merchantProductGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/merchant-products")
	})

	uploadGroup := router.Group("/upload-merchant")
	uploadGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequest(c, service.URL)
	})

	uploadGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequest(c, service.URL)
	})
}

func setupTransactionRoutes(router fiber.Router, service ServiceConfig) {
	transactionGroup := router.Group("/transactions")

	transactionGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/transactions")
	})

	transactionGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/transactions")
	})

	dashboardGroup := router.Group("/dashboard")

	dashboardGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/dashboard")
	})

	dashboardGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/dashboard")
	})
}

func setupWarehouseRoutes(router fiber.Router, service ServiceConfig) {
	warehouseGroup := router.Group("/warehouses")

	warehouseGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/warehouses")
	})

	warehouseGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/warehouses")
	})

	warehouseProductGroup := router.Group("/warehouse-products")
	warehouseProductGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/warehouse-products")
	})

	warehouseProductGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequestWithPath(c, service.URL, "/api/v1/warehouse-products")
	})

	uploadWarehouseGroup := router.Group("/upload-warehouse")
	uploadWarehouseGroup.All("/*", func(c *fiber.Ctx) error {
		return proxyRequest(c, service.URL)
	})

	uploadWarehouseGroup.All("/", func(c *fiber.Ctx) error {
		return proxyRequest(c, service.URL)
	})
}
