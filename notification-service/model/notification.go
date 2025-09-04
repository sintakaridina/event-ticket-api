package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

// NotificationStatus represents the status of a notification
type NotificationStatus string

// NotificationChannel represents the channel through which a notification is sent
type NotificationChannel string

// Notification types
const (
	NotificationTypeBookingConfirmation NotificationType = "booking_confirmation"
	NotificationTypePaymentConfirmation NotificationType = "payment_confirmation"
	NotificationTypePaymentFailed       NotificationType = "payment_failed"
	NotificationTypeEventReminder       NotificationType = "event_reminder"
	NotificationTypeEventCancelled      NotificationType = "event_cancelled"
	NotificationTypeBookingCancelled    NotificationType = "booking_cancelled"
	NotificationTypeRefundProcessed     NotificationType = "refund_processed"
	NotificationTypeCustom              NotificationType = "custom"
)

// Notification statuses
const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// Notification channels
const (
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelPush  NotificationChannel = "push"
)

// Notification represents a notification in the system
type Notification struct {
	ID        uuid.UUID           `gorm:"type:uuid;primary_key" json:"id"`
	UserID    uuid.UUID           `gorm:"type:uuid;index" json:"user_id"`
	Type      NotificationType    `gorm:"type:varchar(50);index" json:"type"`
	Channel   NotificationChannel `gorm:"type:varchar(20);index" json:"channel"`
	Subject   string              `gorm:"type:varchar(255)" json:"subject"`
	Content   string              `gorm:"type:text" json:"content"`
	Status    NotificationStatus  `gorm:"type:varchar(20);index" json:"status"`
	Recipient string              `gorm:"type:varchar(255)" json:"recipient"`
	Metadata  string              `gorm:"type:jsonb" json:"metadata"`
	SentAt    *time.Time          `json:"sent_at"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// NotificationTemplate represents a template for notifications
type NotificationTemplate struct {
	ID        uuid.UUID        `gorm:"type:uuid;primary_key" json:"id"`
	Name      string           `gorm:"type:varchar(100);uniqueIndex" json:"name"`
	Type      NotificationType `gorm:"type:varchar(50);index" json:"type"`
	Channel   NotificationChannel `gorm:"type:varchar(20);index" json:"channel"`
	Subject   string           `gorm:"type:varchar(255)" json:"subject"`
	Content   string           `gorm:"type:text" json:"content"`
	IsActive  bool             `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (nt *NotificationTemplate) BeforeCreate(tx *gorm.DB) error {
	if nt.ID == uuid.Nil {
		nt.ID = uuid.New()
	}
	return nil
}

// NotificationResponse represents the response for a notification
type NotificationResponse struct {
	ID        uuid.UUID           `json:"id"`
	UserID    uuid.UUID           `json:"user_id"`
	Type      NotificationType    `json:"type"`
	Channel   NotificationChannel `json:"channel"`
	Subject   string              `json:"subject"`
	Content   string              `json:"content,omitempty"`
	Status    NotificationStatus  `json:"status"`
	Recipient string              `json:"recipient"`
	SentAt    *time.Time          `json:"sent_at,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
}

// ToResponse converts a Notification to a NotificationResponse
func (n *Notification) ToResponse() NotificationResponse {
	return NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Channel:   n.Channel,
		Subject:   n.Subject,
		Content:   n.Content,
		Status:    n.Status,
		Recipient: n.Recipient,
		SentAt:    n.SentAt,
		CreatedAt: n.CreatedAt,
	}
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID    uuid.UUID           `json:"user_id" binding:"required"`
	Type      NotificationType    `json:"type" binding:"required"`
	Channel   NotificationChannel `json:"channel" binding:"required"`
	Subject   string              `json:"subject" binding:"required"`
	Content   string              `json:"content" binding:"required"`
	Recipient string              `json:"recipient" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SendNotificationRequest represents a request to send a notification
type SendNotificationRequest struct {
	UserID    uuid.UUID           `json:"user_id" binding:"required"`
	Type      NotificationType    `json:"type" binding:"required"`
	Channel   NotificationChannel `json:"channel" binding:"required"`
	TemplateID *uuid.UUID         `json:"template_id,omitempty"`
	TemplateName string           `json:"template_name,omitempty"`
	Subject   string              `json:"subject,omitempty"`
	Content   string              `json:"content,omitempty"`
	Recipient string              `json:"recipient" binding:"required"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// UpdateNotificationStatusRequest represents a request to update a notification status
type UpdateNotificationStatusRequest struct {
	Status NotificationStatus `json:"status" binding:"required"`
}

// CreateTemplateRequest represents a request to create a notification template
type CreateTemplateRequest struct {
	Name      string           `json:"name" binding:"required"`
	Type      NotificationType `json:"type" binding:"required"`
	Channel   NotificationChannel `json:"channel" binding:"required"`
	Subject   string           `json:"subject" binding:"required"`
	Content   string           `json:"content" binding:"required"`
	IsActive  bool             `json:"is_active"`
}

// UpdateTemplateRequest represents a request to update a notification template
type UpdateTemplateRequest struct {
	Name      string           `json:"name,omitempty"`
	Type      NotificationType `json:"type,omitempty"`
	Channel   NotificationChannel `json:"channel,omitempty"`
	Subject   string           `json:"subject,omitempty"`
	Content   string           `json:"content,omitempty"`
	IsActive  *bool            `json:"is_active,omitempty"`
}

// TemplateResponse represents the response for a notification template
type TemplateResponse struct {
	ID        uuid.UUID           `json:"id"`
	Name      string              `json:"name"`
	Type      NotificationType    `json:"type"`
	Channel   NotificationChannel `json:"channel"`
	Subject   string              `json:"subject"`
	Content   string              `json:"content,omitempty"`
	IsActive  bool                `json:"is_active"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// ToResponse converts a NotificationTemplate to a TemplateResponse
func (nt *NotificationTemplate) ToResponse() TemplateResponse {
	return TemplateResponse{
		ID:        nt.ID,
		Name:      nt.Name,
		Type:      nt.Type,
		Channel:   nt.Channel,
		Subject:   nt.Subject,
		Content:   nt.Content,
		IsActive:  nt.IsActive,
		CreatedAt: nt.CreatedAt,
		UpdatedAt: nt.UpdatedAt,
	}
}