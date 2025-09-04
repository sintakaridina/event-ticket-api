package repository

import (
	"github.com/google/uuid"
	"github.com/yourusername/ticket-system/event-ticket-service/model"
	"gorm.io/gorm"
)

// TicketRepository defines the interface for ticket repository operations
type TicketRepository interface {
	Create(ticket *model.Ticket) error
	CreateBatch(tickets []*model.Ticket) error
	FindByID(id uuid.UUID) (*model.Ticket, error)
	FindByEventID(eventID uuid.UUID) ([]model.Ticket, error)
	FindAvailableByEventID(eventID uuid.UUID, ticketType string) ([]model.Ticket, error)
	FindByBookingID(bookingID uuid.UUID) ([]model.Ticket, error)
	Update(ticket *model.Ticket) error
	UpdateBatch(tickets []*model.Ticket) error
	Delete(id uuid.UUID) error
}

// ticketRepository implements TicketRepository interface
type ticketRepository struct {
	db *gorm.DB
}

// NewTicketRepository creates a new ticket repository
func NewTicketRepository(db *gorm.DB) TicketRepository {
	// Auto migrate the models
	db.AutoMigrate(&model.Ticket{})

	return &ticketRepository{
		db: db,
	}
}

// Create creates a new ticket
func (r *ticketRepository) Create(ticket *model.Ticket) error {
	return r.db.Create(ticket).Error
}

// CreateBatch creates multiple tickets in a batch
func (r *ticketRepository) CreateBatch(tickets []*model.Ticket) error {
	return r.db.Create(tickets).Error
}

// FindByID finds a ticket by ID
func (r *ticketRepository) FindByID(id uuid.UUID) (*model.Ticket, error) {
	var ticket model.Ticket
	result := r.db.First(&ticket, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &ticket, nil
}

// FindByEventID finds tickets by event ID
func (r *ticketRepository) FindByEventID(eventID uuid.UUID) ([]model.Ticket, error) {
	var tickets []model.Ticket
	result := r.db.Where("event_id = ?", eventID).Find(&tickets)
	if result.Error != nil {
		return nil, result.Error
	}
	return tickets, nil
}

// FindAvailableByEventID finds available tickets by event ID and type
func (r *ticketRepository) FindAvailableByEventID(eventID uuid.UUID, ticketType string) ([]model.Ticket, error) {
	var tickets []model.Ticket
	query := r.db.Where("event_id = ? AND status = 'available'", eventID)
	if ticketType != "" {
		query = query.Where("type = ?", ticketType)
	}
	result := query.Find(&tickets)
	if result.Error != nil {
		return nil, result.Error
	}
	return tickets, nil
}

// FindByBookingID finds tickets by booking ID
func (r *ticketRepository) FindByBookingID(bookingID uuid.UUID) ([]model.Ticket, error) {
	var tickets []model.Ticket
	result := r.db.Where("booking_id = ?", bookingID).Find(&tickets)
	if result.Error != nil {
		return nil, result.Error
	}
	return tickets, nil
}

// Update updates a ticket
func (r *ticketRepository) Update(ticket *model.Ticket) error {
	return r.db.Save(ticket).Error
}

// UpdateBatch updates multiple tickets in a batch
func (r *ticketRepository) UpdateBatch(tickets []*model.Ticket) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, ticket := range tickets {
			if err := tx.Save(ticket).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete deletes a ticket
func (r *ticketRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Ticket{}, "id = ?", id).Error
}