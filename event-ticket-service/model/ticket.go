package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Ticket represents a ticket for an event
type Ticket struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	EventID   uuid.UUID `gorm:"type:uuid;not null" json:"event_id"`
	Type      string    `gorm:"size:100;not null" json:"type"`
	Price     float64   `gorm:"not null" json:"price"`
	Status    string    `gorm:"size:50;not null;default:'available'" json:"status"` // available, reserved, sold, cancelled
	UserID    uuid.UUID `gorm:"type:uuid" json:"user_id"`
	BookingID uuid.UUID `gorm:"type:uuid" json:"booking_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TicketResponse is the response format for tickets
type TicketResponse struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	Type      string    `json:"type"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	BookingID uuid.UUID `json:"booking_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a Ticket to TicketResponse
func (t *Ticket) ToResponse() TicketResponse {
	return TicketResponse{
		ID:        t.ID,
		EventID:   t.EventID,
		Type:      t.Type,
		Price:     t.Price,
		Status:    t.Status,
		UserID:    t.UserID,
		BookingID: t.BookingID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// Booking represents a booking of tickets
type Booking struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	EventID    uuid.UUID `gorm:"type:uuid;not null" json:"event_id"`
	Status     string    `gorm:"size:50;not null;default:'pending'" json:"status"` // pending, confirmed, cancelled, refunded
	TotalPrice float64   `gorm:"not null" json:"total_price"`
	PaymentID  uuid.UUID `gorm:"type:uuid" json:"payment_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Tickets    []Ticket  `gorm:"foreignKey:BookingID" json:"tickets,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// BookingResponse is the response format for bookings
type BookingResponse struct {
	ID         uuid.UUID        `json:"id"`
	UserID     uuid.UUID        `json:"user_id"`
	EventID    uuid.UUID        `json:"event_id"`
	Event      *EventResponse   `json:"event,omitempty"`
	Status     string           `json:"status"`
	TotalPrice float64          `json:"total_price"`
	PaymentID  uuid.UUID        `json:"payment_id,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	Tickets    []TicketResponse `json:"tickets,omitempty"`
}

// ToResponse converts a Booking to BookingResponse
func (b *Booking) ToResponse(includeEvent bool) BookingResponse {
	response := BookingResponse{
		ID:         b.ID,
		UserID:     b.UserID,
		EventID:    b.EventID,
		Status:     b.Status,
		TotalPrice: b.TotalPrice,
		PaymentID:  b.PaymentID,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}

	// Include tickets if available
	if len(b.Tickets) > 0 {
		tickets := make([]TicketResponse, len(b.Tickets))
		for i, ticket := range b.Tickets {
			tickets[i] = ticket.ToResponse()
		}
		response.Tickets = tickets
	}

	return response
}

// CreateBookingRequest is the request format for creating a booking
type CreateBookingRequest struct {
	EventID uuid.UUID `json:"event_id" binding:"required"`
	Tickets []struct {
		Type     string `json:"type" binding:"required"`
		Quantity int    `json:"quantity" binding:"required"`
	} `json:"tickets" binding:"required"`
}

// UpdateBookingStatusRequest is the request format for updating a booking status
type UpdateBookingStatusRequest struct {
	Status string `json:"status" binding:"required"`
}