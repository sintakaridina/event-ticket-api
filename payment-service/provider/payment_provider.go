package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/ticket-system/payment-service/model"
)

// PaymentProvider defines the interface for payment provider operations
type PaymentProvider interface {
	ProcessPayment(req model.ProcessPaymentRequest) (string, error)
	RefundPayment(transactionID, reason string) error
	VerifyPayment(transactionID string) (bool, error)
}

// paymentProviderType represents the type of payment provider
type paymentProviderType string

const (
	// StripeProvider represents the Stripe payment provider
	StripeProvider paymentProviderType = "stripe"
	// PayPalProvider represents the PayPal payment provider
	PayPalProvider paymentProviderType = "paypal"
	// MockProvider represents a mock payment provider for testing
	MockProvider paymentProviderType = "mock"
)

// NewPaymentProvider creates a new payment provider based on the provider type
func NewPaymentProvider() (PaymentProvider, error) {
	providerType := os.Getenv("PAYMENT_PROVIDER")
	if providerType == "" {
		providerType = string(MockProvider)
	}

	switch paymentProviderType(providerType) {
	case StripeProvider:
		return NewStripeProvider(), nil
	case PayPalProvider:
		return NewPayPalProvider(), nil
	case MockProvider:
		return NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", providerType)
	}
}

// MockProvider implements PaymentProvider interface for testing
type mockProvider struct{}

// NewMockProvider creates a new mock payment provider
func NewMockProvider() PaymentProvider {
	return &mockProvider{}
}

// ProcessPayment processes a payment with the mock provider
func (p *mockProvider) ProcessPayment(req model.ProcessPaymentRequest) (string, error) {
	// Simulate payment processing
	logrus.Info("Processing payment with mock provider")
	logrus.Infof("Amount: %f %s", req.Amount, req.Currency)
	logrus.Infof("Payment method: %s", req.PaymentMethod)

	// Simulate payment validation
	if req.Amount <= 0 {
		return "", errors.New("invalid amount")
	}

	if req.Currency == "" {
		return "", errors.New("invalid currency")
	}

	if req.PaymentMethod == "" {
		return "", errors.New("invalid payment method")
	}

	// Simulate card validation for credit card payments
	if req.PaymentMethod == "credit_card" {
		if req.CardNumber == "" || req.CardExpiry == "" || req.CardCVC == "" || req.CardHolder == "" {
			return "", errors.New("invalid card details")
		}

		// Simulate card validation
		if req.CardNumber == "4111111111111111" {
			return "", errors.New("card declined")
		}
	}

	// Simulate processing delay
	time.Sleep(500 * time.Millisecond)

	// Generate a mock transaction ID
	transactionID := fmt.Sprintf("mock_%d", time.Now().UnixNano())

	logrus.Infof("Payment processed successfully with transaction ID: %s", transactionID)

	return transactionID, nil
}

// RefundPayment refunds a payment with the mock provider
func (p *mockProvider) RefundPayment(transactionID, reason string) error {
	// Simulate refund processing
	logrus.Info("Processing refund with mock provider")
	logrus.Infof("Transaction ID: %s", transactionID)
	logrus.Infof("Reason: %s", reason)

	// Validate transaction ID
	if transactionID == "" {
		return errors.New("invalid transaction ID")
	}

	// Simulate processing delay
	time.Sleep(500 * time.Millisecond)

	logrus.Infof("Refund processed successfully for transaction ID: %s", transactionID)

	return nil
}

// VerifyPayment verifies a payment with the mock provider
func (p *mockProvider) VerifyPayment(transactionID string) (bool, error) {
	// Simulate payment verification
	logrus.Info("Verifying payment with mock provider")
	logrus.Infof("Transaction ID: %s", transactionID)

	// Validate transaction ID
	if transactionID == "" {
		return false, errors.New("invalid transaction ID")
	}

	// Simulate processing delay
	time.Sleep(300 * time.Millisecond)

	// Simulate verification result
	verified := true
	if transactionID == "mock_invalid" {
		verified = false
	}

	logrus.Infof("Payment verification result for transaction ID %s: %t", transactionID, verified)

	return verified, nil
}

// stripeProvider implements PaymentProvider interface for Stripe
type stripeProvider struct {
	secretKey string
	apiURL    string
}

// NewStripeProvider creates a new Stripe payment provider
func NewStripeProvider() PaymentProvider {
	return &stripeProvider{
		secretKey: os.Getenv("STRIPE_SECRET_KEY"),
		apiURL:    "https://api.stripe.com/v1",
	}
}

// ProcessPayment processes a payment with Stripe
func (p *stripeProvider) ProcessPayment(req model.ProcessPaymentRequest) (string, error) {
	if p.secretKey == "" {
		return "", errors.New("STRIPE_SECRET_KEY is not set")
	}

	// Create payment intent payload
	payload := map[string]interface{}{
		"amount":   int(req.Amount * 100), // Convert to cents
		"currency": req.Currency,
		"metadata": map[string]string{
			"user_id":    req.UserID.String(),
			"booking_id": req.BookingID.String(),
		},
		"confirm": true,
		"payment_method": req.PaymentMethodID,
		"return_url": "https://your-website.com/return",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req_http, err := http.NewRequest("POST", p.apiURL+"/payment_intents", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req_http.Header.Set("Authorization", "Bearer "+p.secretKey)
	req_http.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req_http)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("stripe error: %v", result)
	}

	paymentIntentID, ok := result["id"].(string)
	if !ok {
		return "", errors.New("invalid payment intent ID in response")
	}

	logrus.Infof("Stripe payment processed successfully: %s", paymentIntentID)
	return paymentIntentID, nil
}

// RefundPayment refunds a payment with Stripe
func (p *stripeProvider) RefundPayment(transactionID, reason string) error {
	if p.secretKey == "" {
		return errors.New("STRIPE_SECRET_KEY is not set")
	}

	// Create refund payload
	payload := map[string]interface{}{
		"payment_intent": transactionID,
		"reason":         reason,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.apiURL+"/refunds", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("stripe refund error: %v", result)
	}

	logrus.Infof("Stripe refund processed successfully for transaction: %s", transactionID)
	return nil
}

// VerifyPayment verifies a payment with Stripe
func (p *stripeProvider) VerifyPayment(transactionID string) (bool, error) {
	if p.secretKey == "" {
		return false, errors.New("STRIPE_SECRET_KEY is not set")
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", p.apiURL+"/payment_intents/"+transactionID, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+p.secretKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("stripe verification error: %v", result)
	}

	// Check payment status
	status, ok := result["status"].(string)
	if !ok {
		return false, errors.New("invalid status in response")
	}

	isSuccessful := status == "succeeded"
	logrus.Infof("Stripe payment verification for %s: %v", transactionID, isSuccessful)
	return isSuccessful, nil
}

// PayPalProvider implements PaymentProvider interface for PayPal
type paypalProvider struct {
	clientID     string
	clientSecret string
	apiURL       string
	accessToken  string
}

// NewPayPalProvider creates a new PayPal payment provider
func NewPayPalProvider() PaymentProvider {
	return &paypalProvider{
		clientID:     os.Getenv("PAYPAL_CLIENT_ID"),
		clientSecret: os.Getenv("PAYPAL_CLIENT_SECRET"),
		apiURL:       "https://api-m.sandbox.paypal.com", // Use sandbox for development
	}
}

// ProcessPayment processes a payment with PayPal
func (p *paypalProvider) ProcessPayment(req model.ProcessPaymentRequest) (string, error) {
	if p.clientID == "" || p.clientSecret == "" {
		return "", errors.New("PAYPAL_CLIENT_ID or PAYPAL_CLIENT_SECRET is not set")
	}

	// Get access token first
	if err := p.getAccessToken(); err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	// Create order payload
	payload := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": req.Currency,
					"value":         fmt.Sprintf("%.2f", req.Amount),
				},
				"custom_id": req.BookingID.String(),
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req_http, err := http.NewRequest("POST", p.apiURL+"/v2/checkout/orders", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req_http.Header.Set("Authorization", "Bearer "+p.accessToken)
	req_http.Header.Set("Content-Type", "application/json")
	req_http.Header.Set("PayPal-Request-Id", req.BookingID.String())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req_http)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("paypal error: %v", result)
	}

	orderID, ok := result["id"].(string)
	if !ok {
		return "", errors.New("invalid order ID in response")
	}

	logrus.Infof("PayPal payment processed successfully: %s", orderID)
	return orderID, nil
}

// RefundPayment refunds a payment with PayPal
func (p *paypalProvider) RefundPayment(transactionID, reason string) error {
	if p.clientID == "" || p.clientSecret == "" {
		return errors.New("PAYPAL_CLIENT_ID or PAYPAL_CLIENT_SECRET is not set")
	}

	// Get access token first
	if err := p.getAccessToken(); err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Create refund payload
	payload := map[string]interface{}{
		"note_to_payer": reason,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.apiURL+"/v2/payments/captures/"+transactionID+"/refund", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("paypal refund error: %v", result)
	}

	logrus.Infof("PayPal refund processed successfully for transaction: %s", transactionID)
	return nil
}

// VerifyPayment verifies a payment with PayPal
func (p *paypalProvider) VerifyPayment(transactionID string) (bool, error) {
	if p.clientID == "" || p.clientSecret == "" {
		return false, errors.New("PAYPAL_CLIENT_ID or PAYPAL_CLIENT_SECRET is not set")
	}

	// Get access token first
	if err := p.getAccessToken(); err != nil {
		return false, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", p.apiURL+"/v2/checkout/orders/"+transactionID, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("paypal verification error: %v", result)
	}

	// Check payment status
	status, ok := result["status"].(string)
	if !ok {
		return false, errors.New("invalid status in response")
	}

	isSuccessful := status == "COMPLETED"
	logrus.Infof("PayPal payment verification for %s: %v", transactionID, isSuccessful)
	return isSuccessful, nil
}

// getAccessToken gets an access token from PayPal
func (p *paypalProvider) getAccessToken() error {
	// Create request payload
	payload := "grant_type=client_credentials"

	// Create HTTP request
	req, err := http.NewRequest("POST", p.apiURL+"/v1/oauth2/token", bytes.NewBufferString(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(p.clientID, p.clientSecret)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("paypal auth error: %v", result)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return errors.New("invalid access token in response")
	}

	p.accessToken = accessToken
	return nil
}