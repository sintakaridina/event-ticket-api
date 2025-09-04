package utils

import (
	"regexp"
	"time"
)

// IsValidEmail validates an email address
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhoneNumber validates a phone number
func IsValidPhoneNumber(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return phoneRegex.MatchString(phone)
}

// IsValidEventDate validates event dates
func IsValidEventDate(startDate, endDate time.Time) bool {
	// Start date must be in the future
	if startDate.Before(time.Now()) {
		return false
	}

	// End date must be after start date
	if endDate.Before(startDate) {
		return false
	}

	return true
}

// IsValidTicketPrice validates a ticket price
func IsValidTicketPrice(price float64) bool {
	return price >= 0
}

// IsValidEventStatus validates an event status
func IsValidEventStatus(status string) bool {
	validStatuses := map[string]bool{
		"draft":    true,
		"active":   true,
		"cancelled": true,
		"completed": true,
	}

	return validStatuses[status]
}

// IsValidBookingStatus validates a booking status
func IsValidBookingStatus(status string) bool {
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"cancelled": true,
		"refunded":  true,
	}

	return validStatuses[status]
}

// IsValidTicketStatus validates a ticket status
func IsValidTicketStatus(status string) bool {
	validStatuses := map[string]bool{
		"available": true,
		"reserved":  true,
		"sold":      true,
	}

	return validStatuses[status]
}

// IsValidTicketType validates a ticket type
func IsValidTicketType(ticketType string) bool {
	if ticketType == "" {
		return false
	}

	return true
}

// IsValidEventCategory validates an event category
func IsValidEventCategory(category string) bool {
	validCategories := map[string]bool{
		"concert":    true,
		"conference": true,
		"exhibition": true,
		"sports":     true,
		"theater":    true,
		"workshop":   true,
		"other":      true,
	}

	return validCategories[category]
}

// IsValidPagination validates pagination parameters
func IsValidPagination(page, pageSize int) bool {
	return page > 0 && pageSize > 0 && pageSize <= 100
}