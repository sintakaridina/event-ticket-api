package utils

import (
	"fmt"
	"regexp"
	"strings"

	"notification-service/model"
)

var (
	// Email regex pattern
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Phone number regex pattern (international format)
	phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePhone validates a phone number
func ValidatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number is required")
	}

	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone number format, must be in international format (e.g., +1234567890)")
	}

	return nil
}

// ValidateNotificationType validates a notification type
func ValidateNotificationType(notificationType string) error {
	if notificationType == "" {
		return fmt.Errorf("notification type is required")
	}

	validTypes := []string{
		model.NotificationTypePayment,
		model.NotificationTypeTicket,
		model.NotificationTypeUser,
		model.NotificationTypeReminder,
		model.NotificationTypeMarketing,
		model.NotificationTypeSystem,
	}

	for _, validType := range validTypes {
		if notificationType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid notification type: %s", notificationType)
}

// ValidateNotificationChannel validates a notification channel
func ValidateNotificationChannel(channel string) error {
	if channel == "" {
		return fmt.Errorf("notification channel is required")
	}

	validChannels := []string{
		model.NotificationChannelEmail,
		model.NotificationChannelSMS,
		model.NotificationChannelPush,
	}

	for _, validChannel := range validChannels {
		if channel == validChannel {
			return nil
		}
	}

	return fmt.Errorf("invalid notification channel: %s", channel)
}

// ValidateNotificationStatus validates a notification status
func ValidateNotificationStatus(status string) error {
	if status == "" {
		return fmt.Errorf("notification status is required")
	}

	validStatuses := []string{
		model.NotificationStatusPending,
		model.NotificationStatusSending,
		model.NotificationStatusSent,
		model.NotificationStatusFailed,
		model.NotificationStatusRead,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid notification status: %s", status)
}

// ValidateTemplateCode validates a template code
func ValidateTemplateCode(code string) error {
	if code == "" {
		return fmt.Errorf("template code is required")
	}

	// Template code should be snake_case
	snakeCaseRegex := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !snakeCaseRegex.MatchString(code) {
		return fmt.Errorf("template code must be in snake_case format (lowercase letters, numbers, and underscores)")
	}

	return nil
}

// ValidateTemplateContent validates template content for variables
func ValidateTemplateContent(content string, variables map[string]string) error {
	if content == "" {
		return fmt.Errorf("template content is required")
	}

	// Find all variables in the template content
	varRegex := regexp.MustCompile(`{{([^{}]+)}}`)
	matches := varRegex.FindAllStringSubmatch(content, -1)

	// Check if all variables in the template are provided
	for _, match := range matches {
		varName := strings.TrimSpace(match[1])
		if _, exists := variables[varName]; !exists {
			return fmt.Errorf("variable '%s' is used in the template but not provided", varName)
		}
	}

	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, limit int) (int, int) {
	// Default values
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	return page, limit
}

// ValidateDeviceToken validates a device token for push notifications
func ValidateDeviceToken(token string) error {
	if token == "" {
		return fmt.Errorf("device token is required")
	}

	// Device tokens should be at least 32 characters
	if len(token) < 32 {
		return fmt.Errorf("device token is too short")
	}

	return nil
}

// ValidateCreateNotificationRequest validates a create notification request
func ValidateCreateNotificationRequest(req model.CreateNotificationRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if err := ValidateNotificationType(req.Type); err != nil {
		return err
	}

	if err := ValidateNotificationChannel(req.Channel); err != nil {
		return err
	}

	// If template is not specified, title and content are required
	if req.TemplateID == "" && req.TemplateCode == "" {
		if req.Title == "" {
			return fmt.Errorf("title is required when template is not specified")
		}

		if req.Content == "" {
			return fmt.Errorf("content is required when template is not specified")
		}
	}

	// Validate metadata based on channel
	switch req.Channel {
	case model.NotificationChannelEmail:
		if req.Metadata == nil || req.Metadata["email"] == "" {
			return fmt.Errorf("email address is required in metadata for email notifications")
		}
		if err := ValidateEmail(req.Metadata["email"].(string)); err != nil {
			return err
		}
	case model.NotificationChannelSMS:
		if req.Metadata == nil || req.Metadata["phone"] == "" {
			return fmt.Errorf("phone number is required in metadata for SMS notifications")
		}
		if err := ValidatePhone(req.Metadata["phone"].(string)); err != nil {
			return err
		}
	case model.NotificationChannelPush:
		if req.Metadata == nil || req.Metadata["device_token"] == "" {
			return fmt.Errorf("device token is required in metadata for push notifications")
		}
		if err := ValidateDeviceToken(req.Metadata["device_token"].(string)); err != nil {
			return err
		}
	}

	return nil
}

// ValidateCreateTemplateRequest validates a create template request
func ValidateCreateTemplateRequest(req model.CreateTemplateRequest) error {
	if err := ValidateTemplateCode(req.Code); err != nil {
		return err
	}

	if req.Title == "" {
		return fmt.Errorf("template title is required")
	}

	if req.Content == "" {
		return fmt.Errorf("template content is required")
	}

	return nil
}

// ValidateUpdateTemplateRequest validates an update template request
func ValidateUpdateTemplateRequest(req model.UpdateTemplateRequest) error {
	// At least one field must be provided
	if req.Title == "" && req.Content == "" && req.Description == "" {
		return fmt.Errorf("at least one field (title, content, or description) must be provided")
	}

	return nil
}

// ValidateUpdateNotificationStatusRequest validates an update notification status request
func ValidateUpdateNotificationStatusRequest(req model.UpdateNotificationStatusRequest) error {
	return ValidateNotificationStatus(req.Status)
}