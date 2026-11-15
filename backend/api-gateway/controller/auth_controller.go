package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"micro-warehouse/api-gateway/middleware"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

var errInvalidCredentials = errors.New("invalid email or password")

type AuthController struct {
	userServiceURL string
	jwtConfig      middleware.JWTConfig
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type UserServiceResponse struct {
	Data struct {
		UserID uint     `json:"user_id"`
		Email  string   `json:"email"`
		Role   []string `json:"role"`
	} `json:"data"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Roles string `json:"roles"`
	} `json:"user"`
}

func NewAuthController(userServiceURL string, jwtConfig middleware.JWTConfig) *AuthController {
	return &AuthController{
		userServiceURL: userServiceURL,
		jwtConfig:      jwtConfig,
	}
}

func (a *AuthController) Login(c *fiber.Ctx) error {
	var loginRequest LoginRequest
	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if loginRequest.Email == "" || loginRequest.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Email and password are required",
		})
	}

	loginResp, err := a.forwardLoginRequest(loginRequest)
	if err != nil {
		if errors.Is(err, errInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "Bad Gateway",
			"message": err.Error(),
		})
	}

	token, err := middleware.GenerateJWT(loginResp.UserID, loginResp.Email, loginResp.Role, a.jwtConfig)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	response := AuthResponse{
		Token: token,
		User: struct {
			ID    uint   `json:"id"`
			Email string `json:"email"`
			Roles string `json:"roles"`
		}{
			ID:    loginResp.UserID,
			Email: loginResp.Email,
			Roles: loginResp.Role,
		},
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"data":    response,
	})
}

func (ac *AuthController) forwardLoginRequest(loginReq LoginRequest) (*LoginResponse, error) {
	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", ac.userServiceURL+"/api/v1/auth/login", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gateway", "warehouse-api-gateway")
	req.Header.Set("X-Internal-Request", "true")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
		return nil, errInvalidCredentials
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user-service login failed with status %d", resp.StatusCode)
	}

	var userServiceResp UserServiceResponse
	if err := json.Unmarshal(respBody, &userServiceResp); err != nil {
		return nil, err
	}

	rolesStr := ""
	if len(userServiceResp.Data.Role) > 0 {
		rolesStr = userServiceResp.Data.Role[0]
		for i := 1; i < len(userServiceResp.Data.Role); i++ {
			rolesStr += "," + userServiceResp.Data.Role[i]
		}
	}

	loginResp := LoginResponse{
		UserID: userServiceResp.Data.UserID,
		Email:  userServiceResp.Data.Email,
		Role:   rolesStr,
	}

	return &loginResp, nil
}
