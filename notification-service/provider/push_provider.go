package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	serverKey         string
}

// NewFirebasePushProvider creates a new Firebase push notification provider
func NewFirebasePushProvider() PushProvider {
	return &FirebasePushProvider{
		serviceAccountKey: os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY"),
		projectID:         os.Getenv("FIREBASE_PROJECT_ID"),
		serverKey:         os.Getenv("FIREBASE_SERVER_KEY"),
	}
}

// SendPushNotification sends a push notification using Firebase Cloud Messaging
func (p *FirebasePushProvider) SendPushNotification(deviceToken, title, message string, data map[string]interface{}) error {
	// Validate configuration
	if err := p.validateConfig(); err != nil {
		return err
	}

	// Create FCM payload
	payload := map[string]interface{}{
		"to": deviceToken,
		"notification": map[string]interface{}{
			"title": title,
			"body":  message,
			"sound": "default",
		},
		"data": data,
		"priority": "high",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "key="+p.serverKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return fmt.Errorf("FCM error: %v", result)
	}

	// Check success count
	if success, ok := result["success"].(float64); ok && success > 0 {
		logrus.Infof("[FIREBASE] Push notification sent successfully to device %s", deviceToken)
	} else {
		logrus.Warnf("[FIREBASE] Push notification may have failed for device %s: %v", deviceToken, result)
	}

	return nil
}

// validateConfig validates the Firebase configuration
func (p *FirebasePushProvider) validateConfig() error {
	if p.serverKey == "" {
		return fmt.Errorf("FIREBASE_SERVER_KEY is not set")
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