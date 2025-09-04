package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	SendPushNotification(deviceToken, title, message string, data map[string]interface{}) error
}

// FirebasePushProvider implements PushProvider using Firebase Cloud Messaging
type FirebasePushProvider struct {
	serviceAccountKey string
	projectID         string
}

// NewFirebasePushProvider creates a new Firebase push notification provider
func NewFirebasePushProvider() PushProvider {
	return &FirebasePushProvider{
		serviceAccountKey: os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY"),
		projectID:         os.Getenv("FIREBASE_PROJECT_ID"),
	}
}

// SendPushNotification sends a push notification using Firebase Cloud Messaging
func (p *FirebasePushProvider) SendPushNotification(deviceToken, title, message string, data map[string]interface{}) error {
	// Validate configuration
	if err := p.validateConfig(); err != nil {
		return err
	}

	// In a real implementation, this would use the Firebase Admin SDK to send a push notification
	// For now, we'll just log the message
	logrus.Infof("[FIREBASE] Push notification sent to device %s: %s - %s", deviceToken, title, message)

	// TODO: Implement actual Firebase Cloud Messaging API call
	// Example implementation would use the Firebase Admin SDK

	return nil
}

// validateConfig validates the Firebase configuration
func (p *FirebasePushProvider) validateConfig() error {
	if p.serviceAccountKey == "" {
		return fmt.Errorf("FIREBASE_SERVICE_ACCOUNT_KEY is not set")
	}
	if p.projectID == "" {
		return fmt.Errorf("FIREBASE_PROJECT_ID is not set")
	}
	return nil
}

// MockPushProvider implements PushProvider for testing
type MockPushProvider struct {
	SentNotifications []MockPushNotification
}

// MockPushNotification represents a mock push notification for testing
type MockPushNotification struct {
	DeviceToken string
	Title       string
	Message     string
	Data        map[string]interface{}
}

// NewMockPushProvider creates a new mock push notification provider
func NewMockPushProvider() *MockPushProvider {
	return &MockPushProvider{
		SentNotifications: make([]MockPushNotification, 0),
	}
}

// SendPushNotification sends a push notification (mock implementation)
func (p *MockPushProvider) SendPushNotification(deviceToken, title, message string, data map[string]interface{}) error {
	p.SentNotifications = append(p.SentNotifications, MockPushNotification{
		DeviceToken: deviceToken,
		Title:       title,
		Message:     message,
		Data:        data,
	})
	logrus.Infof("[MOCK] Push notification sent to device %s: %s - %s", deviceToken, title, message)
	return nil
}

// NewPushProvider creates a new push notification provider based on configuration
func NewPushProvider() PushProvider {
	providerType := strings.ToLower(os.Getenv("PUSH_PROVIDER"))

	switch providerType {
	case "firebase":
		return NewFirebasePushProvider()
	case "mock", "":
		logrus.Warn("Using mock push notification provider")
		return NewMockPushProvider()
	default:
		logrus.Warnf("Unknown push provider type '%s', using mock provider", providerType)
		return NewMockPushProvider()
	}
}