package repository

import (
	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/notification-service/model"
	"gorm.io/gorm"
)

// NotificationRepository defines the interface for notification repository operations
type NotificationRepository interface {
	// Notification operations
	Create(notification *model.Notification) error
	FindByID(id uuid.UUID) (*model.Notification, error)
	FindByUserID(userID uuid.UUID, page, pageSize int) ([]*model.Notification, int64, error)
	FindByStatus(status model.NotificationStatus, page, pageSize int) ([]*model.Notification, int64, error)
	Update(notification *model.Notification) error
	Delete(id uuid.UUID) error

	// Template operations
	CreateTemplate(template *model.NotificationTemplate) error
	FindTemplateByID(id uuid.UUID) (*model.NotificationTemplate, error)
	FindTemplateByName(name string) (*model.NotificationTemplate, error)
	FindTemplatesByType(notificationType model.NotificationType) ([]*model.NotificationTemplate, error)
	UpdateTemplate(template *model.NotificationTemplate) error
	DeleteTemplate(id uuid.UUID) error
	FindAllTemplates(page, pageSize int) ([]*model.NotificationTemplate, int64, error)

	// Auto-migration
	AutoMigrate() error
}

// GormNotificationRepository implements NotificationRepository using GORM
type GormNotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &GormNotificationRepository{db: db}
}

// Create creates a new notification
func (r *GormNotificationRepository) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

// FindByID finds a notification by ID
func (r *GormNotificationRepository) FindByID(id uuid.UUID) (*model.Notification, error) {
	var notification model.Notification
	result := r.db.Where("id = ?", id).First(&notification)
	if result.Error != nil {
		return nil, result.Error
	}
	return &notification, nil
}

// FindByUserID finds notifications by user ID with pagination
func (r *GormNotificationRepository) FindByUserID(userID uuid.UUID, page, pageSize int) ([]*model.Notification, int64, error) {
	var notifications []*model.Notification
	var total int64

	// Count total records
	if err := r.db.Model(&model.Notification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * pageSize
	result := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return notifications, total, nil
}

// FindByStatus finds notifications by status with pagination
func (r *GormNotificationRepository) FindByStatus(status model.NotificationStatus, page, pageSize int) ([]*model.Notification, int64, error) {
	var notifications []*model.Notification
	var total int64

	// Count total records
	if err := r.db.Model(&model.Notification{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * pageSize
	result := r.db.Where("status = ?", status).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return notifications, total, nil
}

// Update updates a notification
func (r *GormNotificationRepository) Update(notification *model.Notification) error {
	return r.db.Save(notification).Error
}

// Delete deletes a notification
func (r *GormNotificationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Notification{}, id).Error
}

// CreateTemplate creates a new notification template
func (r *GormNotificationRepository) CreateTemplate(template *model.NotificationTemplate) error {
	return r.db.Create(template).Error
}

// FindTemplateByID finds a notification template by ID
func (r *GormNotificationRepository) FindTemplateByID(id uuid.UUID) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	result := r.db.Where("id = ?", id).First(&template)
	if result.Error != nil {
		return nil, result.Error
	}
	return &template, nil
}

// FindTemplateByName finds a notification template by name
func (r *GormNotificationRepository) FindTemplateByName(name string) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	result := r.db.Where("name = ?", name).First(&template)
	if result.Error != nil {
		return nil, result.Error
	}
	return &template, nil
}

// FindTemplatesByType finds notification templates by type
func (r *GormNotificationRepository) FindTemplatesByType(notificationType model.NotificationType) ([]*model.NotificationTemplate, error) {
	var templates []*model.NotificationTemplate
	result := r.db.Where("type = ? AND is_active = true", notificationType).Find(&templates)
	if result.Error != nil {
		return nil, result.Error
	}
	return templates, nil
}

// UpdateTemplate updates a notification template
func (r *GormNotificationRepository) UpdateTemplate(template *model.NotificationTemplate) error {
	return r.db.Save(template).Error
}

// DeleteTemplate deletes a notification template
func (r *GormNotificationRepository) DeleteTemplate(id uuid.UUID) error {
	return r.db.Delete(&model.NotificationTemplate{}, id).Error
}

// FindAllTemplates finds all notification templates with pagination
func (r *GormNotificationRepository) FindAllTemplates(page, pageSize int) ([]*model.NotificationTemplate, int64, error) {
	var templates []*model.NotificationTemplate
	var total int64

	// Count total records
	if err := r.db.Model(&model.NotificationTemplate{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * pageSize
	result := r.db.Order("name ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&templates)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return templates, total, nil
}

// AutoMigrate automatically migrates the notification models
func (r *GormNotificationRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&model.Notification{}, &model.NotificationTemplate{})
}