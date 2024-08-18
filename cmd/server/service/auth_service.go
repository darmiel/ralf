package service

import (
	"github.com/gofiber/fiber/v2"
)

// AuthService defines the interface for authentication services.
type AuthService interface {
	// Middleware returns a Fiber handler that can be used as a middleware
	// to protect routes that require authentication.
	Middleware() fiber.Handler

	// ValidateToken validates the token and returns user information if valid.
	ValidateToken(token string) (*AuthUser, error)
}

// AuthUser represents a user object returned after a successful authentication.
type AuthUser struct {
	UserID   string
	Email    string
	FullName string
}

// AuthMiddleware is a helper function to create middleware from an AuthService.
func AuthMiddleware(authService AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization token not provided")
		}

		user, err := authService.ValidateToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}

		// Store the user information in context, so it can be accessed in handlers.
		c.Locals("user", user)
		return c.Next()
	}
}
