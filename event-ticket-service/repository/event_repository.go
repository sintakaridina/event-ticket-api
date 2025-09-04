package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/event-ticket-service/model"
	"gorm.io/gorm"
)

// EventRepository defines the interface for event repository operations
type EventRepository interface {
	Create(event *model.Event) error
	FindByID(id uuid.UUID) (*model.Event, error)
	Update(event *model.Event) error
	Delete(id uuid.UUID) error
	Search(keyword, category, location string, startDate, endDate time.Time, page, pageSize int) ([]model.Event, int64, error)
	FindAll(page, pageSize int) ([]model.Event, int64, error)
}

// eventRepository implements EventRepository interface
type eventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *gorm.DB) EventRepository {
	// Auto migrate the models
	db.AutoMigrate(&model.Event{})

	return &eventRepository{
		db: db,
	}
}

// Create creates a new event
func (r *eventRepository) Create(event *model.Event) error {
	return r.db.Create(event).Error
}

// FindByID finds an event by ID
func (r *eventRepository) FindByID(id uuid.UUID) (*model.Event, error) {
	var event model.Event
	result := r.db.Preload("Tickets").First(&event, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &event, nil
}

// Update updates an event
func (r *eventRepository) Update(event *model.Event) error {
	return r.db.Save(event).Error
}

// Delete deletes an event
func (r *eventRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Event{}, "id = ?", id).Error
}

// Search searches for events based on criteria
func (r *eventRepository) Search(keyword, category, location string, startDate, endDate time.Time, page, pageSize int) ([]model.Event, int64, error) {
	var events []model.Event
	var total int64

	// Build query
	query := r.db.Model(&model.Event{})

	// Apply filters
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}
	if !startDate.IsZero() {
		query = query.Where("start_date >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("end_date <= ?", endDate)
	}

	// Count total results
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := query.Preload("Tickets").Offset(offset).Limit(pageSize).Find(&events)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return events, total, nil
}

// FindAll finds all events with pagination
func (r *eventRepository) FindAll(page, pageSize int) ([]model.Event, int64, error) {
	var events []model.Event
	var total int64

	// Count total results
	if err := r.db.Model(&model.Event{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := r.db.Preload("Tickets").Offset(offset).Limit(pageSize).Find(&events)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return events, total, nil
}