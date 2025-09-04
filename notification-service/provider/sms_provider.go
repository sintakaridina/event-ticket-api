package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	SendSMS(to, message string) error
}

// TwilioSMSProvider implements SMSProvider using Twilio
type TwilioSMSProvider struct {
	accountSID string
	authToken  string
	fromNumber string
}

// NewTwilioSMSProvider creates a new Twilio SMS provider
func NewTwilioSMSProvider() SMSProvider {
	return &TwilioSMSProvider{
		accountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		authToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		fromNumber: os.Getenv("TWILIO_FROM_NUMBER"),
	}
}

// SendSMS sends an SMS using Twilio
func (p *TwilioSMSProvider) SendSMS(to, message string) error {
	// Validate configuration
	if err := p.validateConfig(); err != nil {
		return err
	}

	// In a real implementation, this would use the Twilio API to send an SMS
	// For now, we'll just log the message
	logrus.Infof("[TWILIO] SMS sent to %s: %s", to, message)

	// TODO: Implement actual Twilio API call
	// Example implementation would use a Twilio client library

	return nil
}

// validateConfig validates the Twilio configuration
func (p *TwilioSMSProvider) validateConfig() error {
	if p.accountSID == "" {
		return fmt.Errorf("TWILIO_ACCOUNT_SID is not set")
	}
	if p.authToken == "" {
		return fmt.Errorf("TWILIO_AUTH_TOKEN is not set")
	}
	if p.fromNumber == "" {
		return fmt.Errorf("TWILIO_FROM_NUMBER is not set")
	}
	return nil
}

// MockSMSProvider implements SMSProvider for testing
type MockSMSProvider struct {
	SentMessages []MockSMS
}

// MockSMS represents a mock SMS for testing
type MockSMS struct {
	To      string
	Message string
}

// NewMockSMSProvider creates a new mock SMS provider
func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{
		SentMessages: make([]MockSMS, 0),
	}
}

// SendSMS sends an SMS (mock implementation)
func (p *MockSMSProvider) SendSMS(to, message string) error {
	p.SentMessages = append(p.SentMessages, MockSMS{
		To:      to,
		Message: message,
	})
	logrus.Infof("[MOCK] SMS sent to %s: %s", to, message)
	return nil
}

// NewSMSProvider creates a new SMS provider based on configuration
func NewSMSProvider() SMSProvider {
	providerType := strings.ToLower(os.Getenv("SMS_PROVIDER"))

	switch providerType {
	case "twilio":
		return NewTwilioSMSProvider()
	case "mock", "":
		logrus.Warn("Using mock SMS provider")
		return NewMockSMSProvider()
	default:
		logrus.Warnf("Unknown SMS provider type '%s', using mock provider", providerType)
		return NewMockSMSProvider()
	}
}