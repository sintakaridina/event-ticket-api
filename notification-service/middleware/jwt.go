package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the claims in the JWT
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// JWTAuth is a middleware for JWT authentication
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the Authorization header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be 'Bearer {token}'"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// validateToken validates the JWT token
func validateToken(tokenString string) (*JWTClaims, error) {
	// Get the JWT secret key from environment variable
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		logrus.Error("JWT_SECRET environment variable is not set")
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		logrus.Errorf("Failed to parse token: %v", err)
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract the claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if the token is expired
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// GenerateToken generates a new JWT token
func GenerateToken(userID, email, role string) (string, error) {
	// Get the JWT secret key from environment variable
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		logrus.Error("JWT_SECRET environment variable is not set")
		return "", fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	// Get the token expiration time from environment variable
	expirationStr := os.Getenv("JWT_EXPIRATION")
	if expirationStr == "" {
		expirationStr = "24h" // Default to 24 hours
	}

	expiration, err := time.ParseDuration(expirationStr)
	if err != nil {
		logrus.Errorf("Failed to parse JWT_EXPIRATION: %v", err)
		return "", fmt.Errorf("invalid JWT_EXPIRATION format: %v", err)
	}

	// Create the claims
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "notification-service",
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		logrus.Errorf("Failed to sign token: %v", err)
		return "", err
	}

	return tokenString, nil
}