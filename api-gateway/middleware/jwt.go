package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWT middleware for validating JWT tokens
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": err.Error()})
			c.Abort()
			return
		}

		claims, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminOnly middleware for restricting access to admin users only
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "user role not found"})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken extracts the JWT token from the request
func extractToken(c *gin.Context) (string, error) {
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

// validateToken validates the JWT token
func validateToken(tokenString string) (*JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logrus.Warn("JWT_SECRET not set, using default secret")
		secret = "default_secret_key"
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// GenerateToken generates a new JWT token
func GenerateToken(userID, email, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logrus.Warn("JWT_SECRET not set, using default secret")
		secret = "default_secret_key"
	}

	expiration := os.Getenv("JWT_EXPIRATION")
	if expiration == "" {
		expiration = "24h"
	}

	expirationDuration, err := time.ParseDuration(expiration)
	if err != nil {
		logrus.Warnf("Invalid JWT_EXPIRATION %s, defaulting to 24h", expiration)
		expirationDuration = 24 * time.Hour
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "api-gateway"
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}