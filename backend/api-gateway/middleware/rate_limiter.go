package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiterConfig konfigurasi untuk rate limiter
type RateLimiterConfig struct {
	// Global rate limit
	GlobalMax        int
	GlobalExpiration time.Duration

	// Auth rate limit (lebih ketat untuk login/register)
	AuthMax        int
	AuthExpiration time.Duration

	// API rate limit (untuk endpoint protected)
	APIMax        int
	APIExpiration time.Duration
}

// DefaultMemoryRateLimiterConfig konfigurasi default untuk in-memory rate limiter
func DefaultMemoryRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		GlobalMax:        1000, // 1000 request per menit
		GlobalExpiration: 1 * time.Minute,

		AuthMax:        5, // 5 login attempt per menit
		AuthExpiration: 1 * time.Minute,

		APIMax:        100, // 100 API calls per menit
		APIExpiration: 1 * time.Minute,
	}
}

// GlobalRateLimiter middleware untuk rate limiting global
func GlobalRateLimiter(config RateLimiterConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.GlobalMax,
		Expiration: config.GlobalExpiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit berdasarkan IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Too Many Requests",
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": config.GlobalExpiration.Seconds(),
			})
		},
		// Storage akan menggunakan memory default dari Fiber
	})
}

// AuthRateLimiter middleware untuk rate limiting khusus auth endpoints
func AuthRateLimiter(config RateLimiterConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.AuthMax,
		Expiration: config.AuthExpiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "auth:" + c.IP() // Prefix untuk auth rate limit
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Too Many Requests",
				"message":     "Too many authentication attempts. Please try again later.",
				"retry_after": config.AuthExpiration.Seconds(),
			})
		},
		// Storage akan menggunakan memory default dari Fiber
	})
}

// APIRateLimiter middleware untuk rate limiting API endpoints
func APIRateLimiter(config RateLimiterConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.APIMax,
		Expiration: config.APIExpiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit berdasarkan IP + user ID (jika ada)
			if uid, ok := c.Locals("user_id").(uint); ok {
				return fmt.Sprintf("api:%s:%d", c.IP(), uid)
			}
			return "api:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Too Many Requests",
				"message":     "API rate limit exceeded. Please try again later.",
				"retry_after": config.APIExpiration.Seconds(),
			})
		},
		// Storage akan menggunakan memory default dari Fiber
	})
}

// CustomRateLimiter middleware untuk rate limiting dengan konfigurasi custom
func CustomRateLimiter(max int, expiration time.Duration, keyPrefix string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return keyPrefix + ":" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Too Many Requests",
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": expiration.Seconds(),
			})
		},
		// Storage akan menggunakan memory default dari Fiber
	})
}
