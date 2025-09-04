package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"notification-service/config"
	"notification-service/model"
	"notification-service/provider"
	"notification-service/repository"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	CreateNotification(req model.CreateNotificationRequest) (*model.Notification, error)
	SendNotification(notificationID string) error
	GetNotificationByID(id string) (*model.Notification, error)
	GetNotificationsByUserID(userID string, page, limit int) ([]model.Notification, int64, error)
	GetNotificationsByStatus(status model.NotificationStatus, page, limit int) ([]model.Notification, int64, error)
	UpdateNotificationStatus(id string, status model.NotificationStatus) error
	CreateTemplate(req model.CreateTemplateRequest) (*model.NotificationTemplate, error)
	UpdateTemplate(id string, req model.UpdateTemplateRequest) (*model.NotificationTemplate, error)
	GetTemplateByID(id string) (*model.NotificationTemplate, error)
	GetTemplateByCode(code string) (*model.NotificationTemplate, error)
	GetAllTemplates(page, limit int) ([]model.NotificationTemplate, int64, error)
	DeleteTemplate(id string) error
	HandlePaymentEvent(msg []byte) error
	HandleTicketEvent(msg []byte) error
	HandleUserEvent(msg []byte) error
}

// NotificationServiceImpl implements NotificationService
type NotificationServiceImpl struct {
	Repo         repository.NotificationRepository
	RabbitMQ     *config.RabbitMQ
	EmailProvider provider.EmailProvider
	SMSProvider   provider.SMSProvider
	PushProvider  provider.PushProvider
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	repo repository.NotificationRepository,
	rabbitMQ *config.RabbitMQ,
	emailProvider provider.EmailProvider,
	smsProvider provider.SMSProvider,
	pushProvider provider.PushProvider,
) NotificationService {
	return &NotificationServiceImpl{
		Repo:          repo,
		RabbitMQ:      rabbitMQ,
		EmailProvider: emailProvider,
		SMSProvider:   smsProvider,
		PushProvider:  pushProvider,
	}
}

// CreateNotification creates a new notification
func (s *NotificationServiceImpl) CreateNotification(req model.CreateNotificationRequest) (*model.Notification, error) {
	// Validate request
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if req.Type == "" {
		return nil, fmt.Errorf("notification type is required")
	}

	if req.Channel == "" {
		return nil, fmt.Errorf("notification channel is required")
	}

	// If template is specified, get it
	var template *model.NotificationTemplate
	var err error
	if req.TemplateID != "" {
		template, err = s.GetTemplateByID(req.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("template not found: %w", err)
		}
	} else if req.TemplateCode != "" {
		template, err = s.GetTemplateByCode(req.TemplateCode)
		if err != nil {
			return nil, fmt.Errorf("template not found: %w", err)
		}
	}

	// Create notification
	notification := &model.Notification{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Type:      req.Type,
		Channel:   req.Channel,
		Title:     req.Title,
		Content:   req.Content,
		Status:    model.NotificationStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// If template is used, apply it
	if template != nil {
		if notification.Title == "" {
			notification.Title = template.Title
		}
		if notification.Content == "" {
			notification.Content = template.Content
		}
		
		// Apply template variables if provided
		if len(req.Variables) > 0 {
			notification.Title = applyTemplateVariables(notification.Title, req.Variables)
			notification.Content = applyTemplateVariables(notification.Content, req.Variables)
		}
	}

	// Store metadata
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		notification.Metadata = string(metadataBytes)
	}

	// Save to database
	if err := s.Repo.CreateNotification(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	logrus.Infof("Notification created: %s", notification.ID)
	return notification, nil
}

// SendNotification sends a notification
func (s *NotificationServiceImpl) SendNotification(notificationID string) error {
	// Get notification
	notification, err := s.GetNotificationByID(notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	// Update status to sending
	if err := s.UpdateNotificationStatus(notification.ID, model.NotificationStatusSending); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	// Send based on channel
	var sendErr error
	switch notification.Channel {
	case model.NotificationChannelEmail:
		sendErr = s.sendEmailNotification(notification)
	case model.NotificationChannelSMS:
		sendErr = s.sendSMSNotification(notification)
	case model.NotificationChannelPush:
		sendErr = s.sendPushNotification(notification)
	default:
		sendErr = fmt.Errorf("unsupported notification channel: %s", notification.Channel)
	}

	// Update status based on result
	if sendErr != nil {
		logrus.Errorf("Failed to send notification %s: %v", notification.ID, sendErr)
		if err := s.UpdateNotificationStatus(notification.ID, model.NotificationStatusFailed); err != nil {
			logrus.Errorf("Failed to update notification status: %v", err)
		}
		return fmt.Errorf("failed to send notification: %w", sendErr)
	}

	// Update status to sent
	if err := s.UpdateNotificationStatus(notification.ID, model.NotificationStatusSent); err != nil {
		logrus.Errorf("Failed to update notification status: %v", err)
	}

	logrus.Infof("Notification sent: %s", notification.ID)
	return nil
}

// GetNotificationByID gets a notification by ID
func (s *NotificationServiceImpl) GetNotificationByID(id string) (*model.Notification, error) {
	notification, err := s.Repo.FindNotificationByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	return notification, nil
}

// GetNotificationsByUserID gets notifications by user ID
func (s *NotificationServiceImpl) GetNotificationsByUserID(userID string, page, limit int) ([]model.Notification, int64, error) {
	notifications, total, err := s.Repo.FindNotificationsByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}
	return notifications, total, nil
}

// GetNotificationsByStatus gets notifications by status
func (s *NotificationServiceImpl) GetNotificationsByStatus(status model.NotificationStatus, page, limit int) ([]model.Notification, int64, error) {
	notifications, total, err := s.Repo.FindNotificationsByStatus(status, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}
	return notifications, total, nil
}

// UpdateNotificationStatus updates a notification's status
func (s *NotificationServiceImpl) UpdateNotificationStatus(id string, status model.NotificationStatus) error {
	notification, err := s.GetNotificationByID(id)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	notification.Status = status
	notification.UpdatedAt = time.Now()

	if err := s.Repo.UpdateNotification(notification); err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return nil
}

// CreateTemplate creates a new notification template
func (s *NotificationServiceImpl) CreateTemplate(req model.CreateTemplateRequest) (*model.NotificationTemplate, error) {
	// Validate request
	if req.Code == "" {
		return nil, fmt.Errorf("template code is required")
	}

	if req.Title == "" {
		return nil, fmt.Errorf("template title is required")
	}

	if req.Content == "" {
		return nil, fmt.Errorf("template content is required")
	}

	// Check if template with same code already exists
	existing, err := s.GetTemplateByCode(req.Code)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("template with code %s already exists", req.Code)
	}

	// Create template
	template := &model.NotificationTemplate{
		ID:          uuid.New().String(),
		Code:        req.Code,
		Title:       req.Title,
		Content:     req.Content,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.Repo.CreateTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	logrus.Infof("Template created: %s (%s)", template.ID, template.Code)
	return template, nil
}

// UpdateTemplate updates a notification template
func (s *NotificationServiceImpl) UpdateTemplate(id string, req model.UpdateTemplateRequest) (*model.NotificationTemplate, error) {
	// Get template
	template, err := s.GetTemplateByID(id)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Update fields
	if req.Title != "" {
		template.Title = req.Title
	}

	if req.Content != "" {
		template.Content = req.Content
	}

	if req.Description != "" {
		template.Description = req.Description
	}

	template.UpdatedAt = time.Now()

	// Save to database
	if err := s.Repo.UpdateTemplate(template); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	logrus.Infof("Template updated: %s (%s)", template.ID, template.Code)
	return template, nil
}

// GetTemplateByID gets a template by ID
func (s *NotificationServiceImpl) GetTemplateByID(id string) (*model.NotificationTemplate, error) {
	template, err := s.Repo.FindTemplateByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return template, nil
}

// GetTemplateByCode gets a template by code
func (s *NotificationServiceImpl) GetTemplateByCode(code string) (*model.NotificationTemplate, error) {
	template, err := s.Repo.FindTemplateByCode(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return template, nil
}

// GetAllTemplates gets all templates with pagination
func (s *NotificationServiceImpl) GetAllTemplates(page, limit int) ([]model.NotificationTemplate, int64, error) {
	templates, total, err := s.Repo.FindAllTemplates(page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get templates: %w", err)
	}
	return templates, total, nil
}

// DeleteTemplate deletes a template
func (s *NotificationServiceImpl) DeleteTemplate(id string) error {
	// Get template
	template, err := s.GetTemplateByID(id)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Delete from database
	if err := s.Repo.DeleteTemplate(template.ID); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	logrus.Infof("Template deleted: %s (%s)", template.ID, template.Code)
	return nil
}

// HandlePaymentEvent handles payment events from RabbitMQ
func (s *NotificationServiceImpl) HandlePaymentEvent(msg []byte) error {
	// Parse event
	var event map[string]interface{}
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to parse payment event: %w", err)
	}

	// Extract event type
	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("invalid payment event: missing event_type")
	}

	// Handle different event types
	switch eventType {
	case "payment_success":
		return s.handlePaymentSuccess(event)
	case "payment_failed":
		return s.handlePaymentFailed(event)
	case "payment_refunded":
		return s.handlePaymentRefunded(event)
	default:
		logrus.Warnf("Unknown payment event type: %s", eventType)
		return nil
	}
}

// HandleTicketEvent handles ticket events from RabbitMQ
func (s *NotificationServiceImpl) HandleTicketEvent(msg []byte) error {
	// Parse event
	var event map[string]interface{}
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to parse ticket event: %w", err)
	}

	// Extract event type
	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("invalid ticket event: missing event_type")
	}

	// Handle different event types
	switch eventType {
	case "ticket_booked":
		return s.handleTicketBooked(event)
	case "ticket_cancelled":
		return s.handleTicketCancelled(event)
	case "event_reminder":
		return s.handleEventReminder(event)
	default:
		logrus.Warnf("Unknown ticket event type: %s", eventType)
		return nil
	}
}

// HandleUserEvent handles user events from RabbitMQ
func (s *NotificationServiceImpl) HandleUserEvent(msg []byte) error {
	// Parse event
	var event map[string]interface{}
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to parse user event: %w", err)
	}

	// Extract event type
	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("invalid user event: missing event_type")
	}

	// Handle different event types
	switch eventType {
	case "user_registered":
		return s.handleUserRegistered(event)
	case "password_reset":
		return s.handlePasswordReset(event)
	default:
		logrus.Warnf("Unknown user event type: %s", eventType)
		return nil
	}
}

// Private helper methods

// sendEmailNotification sends an email notification
func (s *NotificationServiceImpl) sendEmailNotification(notification *model.Notification) error {
	// Extract recipient email from metadata
	var metadata map[string]interface{}
	if notification.Metadata != "" {
		if err := json.Unmarshal([]byte(notification.Metadata), &metadata); err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	email, ok := metadata["email"].(string)
	if !ok || email == "" {
		return fmt.Errorf("email address not found in metadata")
	}

	// Check if HTML content
	isHTML := false
	if htmlFlag, ok := metadata["is_html"].(bool); ok {
		isHTML = htmlFlag
	}

	// Send email
	var err error
	if isHTML {
		err = s.EmailProvider.SendHTMLEmail(email, notification.Title, notification.Content)
	} else {
		err = s.EmailProvider.SendEmail(email, notification.Title, notification.Content)
	}

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendSMSNotification sends an SMS notification
func (s *NotificationServiceImpl) sendSMSNotification(notification *model.Notification) error {
	// Extract phone number from metadata
	var metadata map[string]interface{}
	if notification.Metadata != "" {
		if err := json.Unmarshal([]byte(notification.Metadata), &metadata); err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	phone, ok := metadata["phone"].(string)
	if !ok || phone == "" {
		return fmt.Errorf("phone number not found in metadata")
	}

	// Send SMS
	if err := s.SMSProvider.SendSMS(phone, notification.Content); err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	return nil
}

// sendPushNotification sends a push notification
func (s *NotificationServiceImpl) sendPushNotification(notification *model.Notification) error {
	// Extract device token from metadata
	var metadata map[string]interface{}
	if notification.Metadata != "" {
		if err := json.Unmarshal([]byte(notification.Metadata), &metadata); err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	deviceToken, ok := metadata["device_token"].(string)
	if !ok || deviceToken == "" {
		return fmt.Errorf("device token not found in metadata")
	}

	// Extract additional data if available
	data := make(map[string]interface{})
	if dataMap, ok := metadata["data"].(map[string]interface{}); ok {
		data = dataMap
	}

	// Send push notification
	if err := s.PushProvider.SendPushNotification(deviceToken, notification.Title, notification.Content, data); err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	return nil
}

// applyTemplateVariables replaces template variables in a string
func applyTemplateVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		result = strings.Replace(result, fmt.Sprintf("{{%s}}", key), value, -1)
	}
	return result
}

// publishNotificationEvent publishes a notification event to RabbitMQ
func (s *NotificationServiceImpl) publishNotificationEvent(eventType string, data map[string]interface{}) error {
	// Create event
	event := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().Unix(),
		"data":       data,
	}

	// Marshal to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to RabbitMQ
	if err := s.RabbitMQ.PublishMessage("notification_events", "", amqp.Publishing{
		ContentType: "application/json",
		Body:        eventBytes,
	}); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Event handlers

func (s *NotificationServiceImpl) handlePaymentSuccess(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment event: missing data")
	}

	userID, _ := data["user_id"].(string)
	bookingID, _ := data["booking_id"].(string)
	amount, _ := data["amount"].(float64)
	currency, _ := data["currency"].(string)
	email, _ := data["email"].(string)

	if userID == "" || bookingID == "" || email == "" {
		return fmt.Errorf("invalid payment event: missing required fields")
	}

	// Create notification using template
	variables := map[string]string{
		"booking_id": bookingID,
		"amount":     fmt.Sprintf("%.2f %s", amount, currency),
	}

	req := model.CreateNotificationRequest{
		UserID:       userID,
		Type:         model.NotificationTypePayment,
		Channel:      model.NotificationChannelEmail,
		TemplateCode: "payment_success",
		Variables:    variables,
		Metadata: map[string]interface{}{
			"email":   email,
			"is_html": true,
		},
	}

	notification, err := s.CreateNotification(req)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := s.SendNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *NotificationServiceImpl) handlePaymentFailed(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment event: missing data")
	}

	userID, _ := data["user_id"].(string)
	bookingID, _ := data["booking_id"].(string)
	reason, _ := data["reason"].(string)
	email, _ := data["email"].(string)

	if userID == "" || bookingID == "" || email == "" {
		return fmt.Errorf("invalid payment event: missing required fields")
	}

	// Create notification using template
	variables := map[string]string{
		"booking_id": bookingID,
		"reason":     reason,
	}

	req := model.CreateNotificationRequest{
		UserID:       userID,
		Type:         model.NotificationTypePayment,
		Channel:      model.NotificationChannelEmail,
		TemplateCode: "payment_failed",
		Variables:    variables,
		Metadata: map[string]interface{}{
			"email":   email,
			"is_html": true,
		},
	}

	notification, err := s.CreateNotification(req)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := s.SendNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *NotificationServiceImpl) handlePaymentRefunded(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment event: missing data")
	}

	userID, _ := data["user_id"].(string)
	bookingID, _ := data["booking_id"].(string)
	amount, _ := data["amount"].(float64)
	currency, _ := data["currency"].(string)
	email, _ := data["email"].(string)

	if userID == "" || bookingID == "" || email == "" {
		return fmt.Errorf("invalid payment event: missing required fields")
	}

	// Create notification using template
	variables := map[string]string{
		"booking_id": bookingID,
		"amount":     fmt.Sprintf("%.2f %s", amount, currency),
	}

	req := model.CreateNotificationRequest{
		UserID:       userID,
		Type:         model.NotificationTypePayment,
		Channel:      model.NotificationChannelEmail,
		TemplateCode: "payment_refunded",
		Variables:    variables,
		Metadata: map[string]interface{}{
			"email":   email,
			"is_html": true,
		},
	}

	notification, err := s.CreateNotification(req)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := s.SendNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *NotificationServiceImpl) handleTicketBooked(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid ticket event: missing data")
	}

	userID, _ := data["user_id"].(string)
	bookingID, _ := data["booking_id"].(string)
	eventName, _ := data["event_name"].(string)
	eventDate, _ := data["event_date"].(string)
	eventLocation, _ := data["event_location"].(string)
	ticketCount, _ := data["ticket_count"].(float64)
	email, _ := data["email"].(string)

	if userID == "" || bookingID == "" || email == "" {
		return fmt.Errorf("invalid ticket event: missing required fields")
	}

	// Create notification using template
	variables := map[string]string{
		"booking_id":     bookingID,
		"event_name":     eventName,
		"event_date":     eventDate,
		"event_location": eventLocation,
		"ticket_count":   fmt.Sprintf("%.0f", ticketCount),
	}

	req := model.CreateNotificationRequest{
		UserID:       userID,
		Type:         model.NotificationTypeTicket,
		Channel:      model.NotificationChannelEmail,
		TemplateCode: "ticket_booked",
		Variables:    variables,
		Metadata: map[string]interface{}{
			"email":   email,
			"is_html": true,
		},
	}

	notification, err := s.CreateNotification(req)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := s.SendNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *NotificationServiceImpl) handleTicketCancelled(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid ticket event: missing data")
	}

	userID, _ := data["user_id"].(string)
	bookingID, _ := data["booking_id"].(string)
	eventName, _ := data["event_name"].(string)
	reason, _ := data["reason"].(string)
	email, _ := data["email"].(string)

	if userID == "" || bookingID == "" || email == "" {
		return fmt.Errorf("invalid ticket event: missing required fields")
	}

	// Create notification using template
	variables := map[string]string{
		"booking_id": bookingID,
		"event_name": eventName,
		"reason":     reason,
	}

	req := model.CreateNotificationRequest{
		UserID:       userID,
		Type:         model.NotificationTypeTicket,
		Channel:      model.NotificationChannelEmail,
		TemplateCode: "ticket_cancelled",
		Variables:    variables,
		Metadata: map[string]interface{}{
			"email":   email,
			"is_html": true,
		},
	}

	notification, err := s.CreateNotification(req)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := s.SendNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *NotificationServiceImpl) handleEventReminder(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid ticket event: missing data")
	}

	userID, _ := data["user_id"].(string)
	eventName, _ := data["event_name"].(string)
	eventDate, _ := data["event_date"].(string)
	eventLocation, _ := data["event_location"].(string)
	email, _ := data["email"].(string)
	phone, _ := data["phone"].(string)

	if userID == "" || eventName == "" || eventDate == "" || (email == "" && phone == "") {
		return fmt.Errorf("invalid event reminder: missing required fields")
	}

	// Create variables
	variables := map[string]string{
		"event_name":     eventName,
		"event_date":     eventDate,
		"event_location": eventLocation,
	}

	// Send email reminder if email is provided
	if email != "" {
		emailReq := model.CreateNotificationRequest{
			UserID:       userID,
			Type:         model.NotificationTypeReminder,
			Channel:      model.NotificationChannelEmail,
			TemplateCode: "event_reminder",
			Variables:    variables,
			Metadata: map[string]interface{}{
				"email":   email,
				"is_html": true,
			},
		}

		notification, err := s.CreateNotification(emailReq)
		if err != nil {
			logrus.Errorf("Failed to create email reminder: %v", err)
		} else {
			if err := s.SendNotification(notification.ID); err != nil {
				logrus.Errorf("Failed to send email reminder: %v", err)
			}
		}
	}

	// Send SMS reminder if phone is provided
	if phone != "" {
		smsReq := model.CreateNotificationRequest{
			UserID:       userID,
			Type:         model.NotificationTypeReminder,
			Channel:      model.NotificationChannelSMS,
			TemplateCode: "event_reminder_sms",
			Variables:    variables,
			Metadata: map[string]interface{}{
				"phone": phone,
			},
		}

		notification, err := s.CreateNotification(smsReq)
		if err != nil {
			logrus.Errorf("Failed to create SMS reminder: %v", err)
		} else {
			if err := s.SendNotification(notification.ID); err != nil {
				logrus.Errorf("Failed to send SMS reminder: %v", err)
			}
		}
	}

	return nil
}

func (s *NotificationServiceImpl) handleUserRegistered(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid user event: missing data")
	}

	userID, _ := data["user_id"].(string)
	username, _ := data["username"].(string)
	email, _ := data["email"].(string)

	if userID == "" || email == "" {
		return fmt.Errorf("invalid user event: missing required fields")
	}

	// Create notification using template
	variables := map[string]string{
		"username": username,
	}

	req := model.CreateNotificationRequest{
		UserID:       userID,
		Type:         model.NotificationTypeUser,
		Channel:      model.NotificationChannelEmail,
		TemplateCode: "welcome_email",
		Variables:    variables,
		Metadata: map[string]interface{}{
			"email":   email,
			"is_html": true,
		},
	}

	notification, err := s.CreateNotification(req)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := s.SendNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *NotificationServiceImpl) handlePasswordReset(event map[string]interface{}) error {
	// Extract data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid user event: missing data")
	}

	userID, _ := data["user_id"].(string)
	email, _ := data["email"].(string)
	resetToken, _ := data["reset_token"].(string)
	resetURL, _ := data["reset_url"].(string)

	if userID == "" || email == "" || (resetToken == "" && resetURL == "") {
		return fmt.Errorf("invalid password reset event: missing required fields")
	}

	// Create reset URL if only token is provided
	if resetURL == "" && resetToken != "" {
		resetURL = fmt.Sprintf("https://example.com/reset-password?token=%s", resetToken)
	}

	// Create notification using template
	variables := map[string]string{
		"reset_url": resetURL,
	}

	req := model.CreateNotificationRequest{
		UserID:       userID,
		Type:         model.NotificationTypeUser,
		Channel:      model.NotificationChannelEmail,
		TemplateCode: "password_reset",
		Variables:    variables,
		Metadata: map[string]interface{}{
			"email":   email,
			"is_html": true,
		},
	}

	notification, err := s.CreateNotification(req)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification
	if err := s.SendNotification(notification.ID); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}