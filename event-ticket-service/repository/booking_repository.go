package repository

import (
	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/event-ticket-service/model"
	"gorm.io/gorm"
)

// BookingRepository defines the interface for booking repository operations
type BookingRepository interface {
	Create(booking *model.Booking) error
	FindByID(id uuid.UUID) (*model.Booking, error)
	FindByUserID(userID uuid.UUID, page, pageSize int) ([]model.Booking, int64, error)
	FindByEventID(eventID uuid.UUID, page, pageSize int) ([]model.Booking, int64, error)
	Update(booking *model.Booking) error
	Delete(id uuid.UUID) error
}

// bookingRepository implements BookingRepository interface
type bookingRepository struct {
	db *gorm.DB
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *gorm.DB) BookingRepository {
	// Auto migrate the models
	db.AutoMigrate(&model.Booking{})

	return &bookingRepository{
		db: db,
	}
}

// Create creates a new booking
func (r *bookingRepository) Create(booking *model.Booking) error {
	return r.db.Create(booking).Error
}

// FindByID finds a booking by ID
func (r *bookingRepository) FindByID(id uuid.UUID) (*model.Booking, error) {
	var booking model.Booking
	result := r.db.Preload("Tickets").First(&booking, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &booking, nil
}

// FindByUserID finds bookings by user ID with pagination
func (r *bookingRepository) FindByUserID(userID uuid.UUID, page, pageSize int) ([]model.Booking, int64, error) {
	var bookings []model.Booking
	var total int64

	// Count total results
	if err := r.db.Model(&model.Booking{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := r.db.Preload("Tickets").Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Find(&bookings)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return bookings, total, nil
}

// FindByEventID finds bookings by event ID with pagination
func (r *bookingRepository) FindByEventID(eventID uuid.UUID, page, pageSize int) ([]model.Booking, int64, error) {
	var bookings []model.Booking
	var total int64

	// Count total results
	if err := r.db.Model(&model.Booking{}).Where("event_id = ?", eventID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	result := r.db.Preload("Tickets").Where("event_id = ?", eventID).Offset(offset).Limit(pageSize).Find(&bookings)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return bookings, total, nil
}

// Update updates a booking
func (r *bookingRepository) Update(booking *model.Booking) error {
	return r.db.Save(booking).Error
}

// Delete deletes a booking
func (r *bookingRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Booking{}, "id = ?", id).Error
}