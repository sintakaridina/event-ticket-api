package provider

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// EmailProvider defines the interface for email providers
type EmailProvider interface {
	SendEmail(to, subject, body string) error
	SendHTMLEmail(to, subject, htmlBody string) error
}

// SMTPEmailProvider implements EmailProvider using SMTP
type SMTPEmailProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewSMTPEmailProvider creates a new SMTP email provider
func NewSMTPEmailProvider() EmailProvider {
	return &SMTPEmailProvider{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

// SendEmail sends a plain text email
func (p *SMTPEmailProvider) SendEmail(to, subject, body string) error {
	// Validate configuration
	if err := p.validateConfig(); err != nil {
		return err
	}

	// Set up authentication information
	auth := smtp.PlainAuth("", p.username, p.password, p.host)

	// Compose message
	message := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n\r\n"+
			"%s\r\n",
		p.from, to, subject, body))

	// Send email
	addr := fmt.Sprintf("%s:%s", p.host, p.port)
	if err := smtp.SendMail(addr, auth, p.from, []string{to}, message); err != nil {
		logrus.Errorf("Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logrus.Infof("Email sent to %s with subject '%s'", to, subject)
	return nil
}

// SendHTMLEmail sends an HTML email
func (p *SMTPEmailProvider) SendHTMLEmail(to, subject, htmlBody string) error {
	// Validate configuration
	if err := p.validateConfig(); err != nil {
		return err
	}

	// Set up authentication information
	auth := smtp.PlainAuth("", p.username, p.password, p.host)

	// Compose message with MIME headers for HTML
	message := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"%s\r\n",
		p.from, to, subject, htmlBody))

	// Send email
	addr := fmt.Sprintf("%s:%s", p.host, p.port)
	if err := smtp.SendMail(addr, auth, p.from, []string{to}, message); err != nil {
		logrus.Errorf("Failed to send HTML email: %v", err)
		return fmt.Errorf("failed to send HTML email: %w", err)
	}

	logrus.Infof("HTML email sent to %s with subject '%s'", to, subject)
	return nil
}

// validateConfig validates the SMTP configuration
func (p *SMTPEmailProvider) validateConfig() error {
	if p.host == "" {
		return fmt.Errorf("SMTP_HOST is not set")
	}
	if p.port == "" {
		return fmt.Errorf("SMTP_PORT is not set")
	}
	if p.username == "" {
		return fmt.Errorf("SMTP_USERNAME is not set")
	}
	if p.password == "" {
		return fmt.Errorf("SMTP_PASSWORD is not set")
	}
	if p.from == "" {
		return fmt.Errorf("SMTP_FROM is not set")
	}
	return nil
}

// MockEmailProvider implements EmailProvider for testing
type MockEmailProvider struct {
	SentEmails []MockEmail
}

// MockEmail represents a mock email for testing
type MockEmail struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

// NewMockEmailProvider creates a new mock email provider
func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{
		SentEmails: make([]MockEmail, 0),
	}
}

// SendEmail sends a plain text email (mock implementation)
func (p *MockEmailProvider) SendEmail(to, subject, body string) error {
	p.SentEmails = append(p.SentEmails, MockEmail{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	})
	logrus.Infof("[MOCK] Email sent to %s with subject '%s'", to, subject)
	return nil
}

// SendHTMLEmail sends an HTML email (mock implementation)
func (p *MockEmailProvider) SendHTMLEmail(to, subject, htmlBody string) error {
	p.SentEmails = append(p.SentEmails, MockEmail{
		To:      to,
		Subject: subject,
		Body:    htmlBody,
		IsHTML:  true,
	})
	logrus.Infof("[MOCK] HTML email sent to %s with subject '%s'", to, subject)
	return nil
}

// NewEmailProvider creates a new email provider based on configuration
func NewEmailProvider() EmailProvider {
	providerType := strings.ToLower(os.Getenv("EMAIL_PROVIDER"))

	switch providerType {
	case "smtp":
		return NewSMTPEmailProvider()
	case "mock", "":
		logrus.Warn("Using mock email provider")
		return NewMockEmailProvider()
	default:
		logrus.Warnf("Unknown email provider type '%s', using mock provider", providerType)
		return NewMockEmailProvider()
	}
}