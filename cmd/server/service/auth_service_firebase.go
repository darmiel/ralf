package service

import (
	"context"
	"fmt"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
)

// make sure FirebaseAuthService implements AuthService
var _ AuthService = (*FirebaseAuthService)(nil)

// FirebaseAuthService is an implementation of AuthService using Firebase.
type FirebaseAuthService struct {
	app *firebase.App
}

// NewFirebaseAuthService creates a new FirebaseAuthService.
func NewFirebaseAuthService(credentialsPath string) (*FirebaseAuthService, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}
	return &FirebaseAuthService{app: app}, nil
}

// Middleware returns a middleware function for Firebase authentication.
func (f *FirebaseAuthService) Middleware() fiber.Handler {
	return AuthMiddleware(f)
}

// ValidateToken validates the Firebase token and extracts user information.
func (f *FirebaseAuthService) ValidateToken(token string) (*AuthUser, error) {
	auth, err := f.app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase auth client: %w", err)
	}

	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	decodedToken, err := auth.VerifyIDToken(context.Background(), token)
	if err != nil {
		return nil, fmt.Errorf("invalid Firebase token: %w", err)
	}

	// Example of extracting user information from the decoded token.
	uid := decodedToken.UID
	email := decodedToken.Claims["email"].(string)

	return &AuthUser{
		UserID: uid,
		Email:  email,
	}, nil
}
