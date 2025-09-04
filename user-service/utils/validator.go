package utils

import (
	"regexp"
	"unicode"
)

// IsValidEmail validates an email address
func IsValidEmail(email string) bool {
	// Simple regex for email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsStrongPassword checks if a password meets security requirements
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (hasUpper, hasLower, hasDigit, hasSpecial bool)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsValidPhone validates a phone number
func IsValidPhone(phone string) bool {
	// Simple regex for phone validation (international format)
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return phoneRegex.MatchString(phone)
}