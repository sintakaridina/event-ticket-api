package utils

import (
	"regexp"
	"strings"
	"time"
)

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

// ValidatePaymentMethod validates a payment method
func ValidatePaymentMethod(method string) bool {
	validMethods := map[string]bool{
		"credit_card": true,
		"debit_card":  true,
		"paypal":      true,
		"bank_transfer": true,
		"crypto":      true,
	}

	return validMethods[strings.ToLower(method)]
}

// ValidatePaymentStatus validates a payment status
func ValidatePaymentStatus(status string) bool {
	validStatuses := map[string]bool{
		"pending":   true,
		"completed": true,
		"failed":    true,
		"refunded":  true,
		"cancelled": true,
	}

	return validStatuses[strings.ToLower(status)]
}

// ValidateCurrency validates a currency code
func ValidateCurrency(currency string) bool {
	validCurrencies := map[string]bool{
		"usd": true,
		"eur": true,
		"gbp": true,
		"jpy": true,
		"cad": true,
		"aud": true,
		"idr": true,
		"sgd": true,
		"myr": true,
	}

	return validCurrencies[strings.ToLower(currency)]
}

// ValidateCardNumber validates a credit card number using Luhn algorithm
func ValidateCardNumber(cardNumber string) bool {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// Check if the card number contains only digits
	if !regexp.MustCompile(`^\d+$`).MatchString(cardNumber) {
		return false
	}

	// Check length (most card numbers are between 13 and 19 digits)
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}

	// Luhn algorithm
	sum := 0
	isSecond := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		d := int(cardNumber[i] - '0')

		if isSecond {
			d *= 2
		}

		sum += d / 10
		sum += d % 10

		isSecond = !isSecond
	}

	return sum%10 == 0
}

// ValidateCardExpiry validates a credit card expiry date (MM/YY format)
func ValidateCardExpiry(expiry string) bool {
	// Check format
	if !regexp.MustCompile(`^\d{2}/\d{2}$`).MatchString(expiry) {
		return false
	}

	// Split into month and year
	parts := strings.Split(expiry, "/")
	monthStr, yearStr := parts[0], parts[1]

	// Parse month and year
	month, err := time.Parse("01", monthStr)
	if err != nil || month.Month() < 1 || month.Month() > 12 {
		return false
	}

	// Parse year (add 2000 to get the full year)
	yearInt := 2000 + int(yearStr[0]-'0')*10 + int(yearStr[1]-'0')
	
	// Get current date
	now := time.Now()
	currentYear, currentMonth := now.Year(), now.Month()

	// Check if the card is expired
	if yearInt < currentYear || (yearInt == currentYear && month.Month() < currentMonth) {
		return false
	}

	return true
}

// ValidateCardCVC validates a credit card CVC/CVV
func ValidateCardCVC(cvc string) bool {
	// CVC/CVV is typically 3 or 4 digits
	return regexp.MustCompile(`^\d{3,4}$`).MatchString(cvc)
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return page, pageSize
}

// ValidateAmount validates a payment amount
func ValidateAmount(amount float64) bool {
	return amount > 0
}

// ValidateRefundReason validates a refund reason
func ValidateRefundReason(reason string) bool {
	return len(reason) > 0 && len(reason) <= 500
}