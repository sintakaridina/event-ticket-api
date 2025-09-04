package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the claims in the JWT
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth is a middleware that validates JWT tokens
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(c *gin.Context) string {
	bearer := c.GetHeader("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:7]) == "BEARER " {
		return bearer[7:]
	}
	return ""
}

// validateToken validates the JWT token
func validateToken(tokenString string) (*JWTClaims, error) {
	// Get JWT secret from environment variable
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logrus.Warn("JWT_SECRET not set, using default secret")
		secret = "default_jwt_secret"
	}

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Validate claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token expired")
		}

		// Check if user ID is present
		if claims.UserID == "" {
			return nil, fmt.Errorf("invalid token: missing user ID")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}