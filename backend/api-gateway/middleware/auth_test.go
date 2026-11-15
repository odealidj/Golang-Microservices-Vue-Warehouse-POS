package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func testJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey: "unit-test-secret",
		Issuer:    "warehouse-test",
		Duration:  time.Hour,
	}
}

func TestJWTAuthMiddleware_RejectsSpoofedInternalHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(JWTAuthMiddleware(testJWTConfig()))
	app.Get("/api/v1/products", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	req.Header.Set("X-Internal-Request", "true")
	req.Header.Set("X-Gateway", "warehouse-api-gateway")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for spoofed internal headers without JWT, got %d", resp.StatusCode)
	}
}

func TestJWTAuthMiddleware_ValidTokenSetsContext(t *testing.T) {
	cfg := testJWTConfig()
	token, err := GenerateJWT(42, "keeper@toko.id", "Keeper", cfg)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	app := fiber.New()
	app.Use(JWTAuthMiddleware(cfg))
	app.Get("/api/v1/transactions", func(c *fiber.Ctx) error {
		if c.Locals("user_id") != uint(42) {
			return c.Status(fiber.StatusInternalServerError).SendString("wrong user_id")
		}
		if c.Locals("user_roles") != "Keeper" {
			return c.Status(fiber.StatusInternalServerError).SendString("wrong roles")
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestJWTAuthMiddleware_PublicRoutesAllowed(t *testing.T) {
	app := fiber.New()
	app.Use(JWTAuthMiddleware(testJWTConfig()))
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Post("/api/v1/midtrans/callback", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
		{http.MethodPost, "/api/v1/midtrans/callback"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test(%s %s) failed: %v", tc.method, tc.path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200 OK for %s %s, got %d", tc.method, tc.path, resp.StatusCode)
		}
	}
}

func TestRoleAuthMiddleware_StringRoles(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_roles", "Keeper, Manager")
		return c.Next()
	})
	app.Get("/admin", RoleAuthMiddleware("manager"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/denied", RoleAuthMiddleware("superadmin"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	respOK, err := app.Test(httptest.NewRequest(http.MethodGet, "/admin", nil))
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer func() { _ = respOK.Body.Close() }()
	if respOK.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respOK.StatusCode)
	}

	respForbidden, err := app.Test(httptest.NewRequest(http.MethodGet, "/denied", nil))
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer func() { _ = respForbidden.Body.Close() }()
	if respForbidden.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", respForbidden.StatusCode)
	}
}
