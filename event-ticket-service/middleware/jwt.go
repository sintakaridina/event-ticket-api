package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the claims in the JWT token
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth is a middleware that validates JWT tokens
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := validateToken(tokenString)
		if err != nil {
			logrus.WithError(err).Warn("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
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

// extractToken extracts the token from the Authorization header
func extractToken(c *gin.Context) string {
	bearer := c.GetHeader("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:7]) == "BEARER " {
		return bearer[7:]
	}
	return ""
}

// validateToken validates the token and returns the claims
func validateToken(tokenString string) (*JWTClaims, error) {
	// Get JWT secret from environment variable
	jwtSecret := getJWTSecret()

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// getJWTSecret gets the JWT secret from environment variable or returns a default
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// In production, you should ensure a proper secret is set
		return "your-default-jwt-secret-for-development-only"
	}
	return secret
}