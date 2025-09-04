package repository

import (
	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/payment-service/model"
	"gorm.io/gorm"
)

// PaymentRepository defines the interface for payment repository operations
type PaymentRepository interface {
	Create(payment *model.Payment) error
	FindByID(id uuid.UUID) (*model.Payment, error)
	FindByBookingID(bookingID uuid.UUID) (*model.Payment, error)
	FindByUserID(userID uuid.UUID, page, pageSize int) ([]model.Payment, int64, error)
	Update(payment *model.Payment) error
	Delete(id uuid.UUID) error
}

// paymentRepositoryImpl implements PaymentRepository interface
type paymentRepositoryImpl struct {
	db *gorm.DB
}

// NewPaymentRepositoryImpl creates a new payment repository
func NewPaymentRepositoryImpl(db *gorm.DB) PaymentRepository {
	// Auto migrate the Payment model
	db.AutoMigrate(&model.Payment{})

	return &paymentRepositoryImpl{
		db: db,
	}
}

// Create creates a new payment
func (r *paymentRepositoryImpl) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

// FindByID finds a payment by ID
func (r *paymentRepositoryImpl) FindByID(id uuid.UUID) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("id = ?", id).First(&payment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &payment, nil
}

// FindByBookingID finds a payment by booking ID
func (r *paymentRepositoryImpl) FindByBookingID(bookingID uuid.UUID) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("booking_id = ?", bookingID).First(&payment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &payment, nil
}

// FindByUserID finds payments by user ID with pagination
func (r *paymentRepositoryImpl) FindByUserID(userID uuid.UUID, page, pageSize int) ([]model.Payment, int64, error) {
	var payments []model.Payment
	var total int64

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	err := r.db.Model(&model.Payment{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get payments with pagination
	err = r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&payments).Error

	if err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// Update updates a payment
func (r *paymentRepositoryImpl) Update(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

// Delete deletes a payment
func (r *paymentRepositoryImpl) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Payment{}).Error
}