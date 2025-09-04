package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Event represents an event in the system
type Event struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Location    string    `gorm:"size:255;not null" json:"location"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`
	Category    string    `gorm:"size:100;not null" json:"category"`
	Organizer   string    `gorm:"size:255;not null" json:"organizer"`
	ImageURL    string    `gorm:"size:255" json:"image_url"`
	Status      string    `gorm:"size:50;not null;default:'active'" json:"status"` // active, cancelled, completed
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Tickets     []Ticket  `gorm:"foreignKey:EventID" json:"tickets,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// EventResponse is the response format for events
type EventResponse struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Location    string         `json:"location"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Category    string         `json:"category"`
	Organizer   string         `json:"organizer"`
	ImageURL    string         `json:"image_url"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Tickets     []TicketType   `json:"tickets,omitempty"`
}

// ToResponse converts an Event to EventResponse
func (e *Event) ToResponse() EventResponse {
	ticketTypes := make([]TicketType, 0)
	ticketMap := make(map[string]TicketType)

	// Group tickets by type
	for _, ticket := range e.Tickets {
		if ticket.Status == "available" {
			if tt, exists := ticketMap[ticket.Type]; exists {
				tt.AvailableQuantity++
				ticketMap[ticket.Type] = tt
			} else {
				ticketMap[ticket.Type] = TicketType{
					Type:              ticket.Type,
					Price:             ticket.Price,
					AvailableQuantity: 1,
				}
			}
		}
	}

	// Convert map to slice
	for _, tt := range ticketMap {
		ticketTypes = append(ticketTypes, tt)
	}

	return EventResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Location:    e.Location,
		StartDate:   e.StartDate,
		EndDate:     e.EndDate,
		Category:    e.Category,
		Organizer:   e.Organizer,
		ImageURL:    e.ImageURL,
		Status:      e.Status,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		Tickets:     ticketTypes,
	}
}

// CreateEventRequest is the request format for creating an event
type CreateEventRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location" binding:"required"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	Category    string    `json:"category" binding:"required"`
	Organizer   string    `json:"organizer" binding:"required"`
	ImageURL    string    `json:"image_url"`
	Tickets     []struct {
		Type     string  `json:"type" binding:"required"`
		Price    float64 `json:"price" binding:"required"`
		Quantity int     `json:"quantity" binding:"required"`
	} `json:"tickets" binding:"required"`
}

// UpdateEventRequest is the request format for updating an event
type UpdateEventRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Category    string    `json:"category"`
	Organizer   string    `json:"organizer"`
	ImageURL    string    `json:"image_url"`
	Status      string    `json:"status"`
}

// SearchEventRequest is the request format for searching events
type SearchEventRequest struct {
	Keyword    string    `form:"keyword"`
	Category   string    `form:"category"`
	Location   string    `form:"location"`
	StartDate  time.Time `form:"start_date"`
	EndDate    time.Time `form:"end_date"`
	Page       int       `form:"page,default=1"`
	PageSize   int       `form:"page_size,default=10"`
}

// TicketType represents a type of ticket for an event
type TicketType struct {
	Type              string  `json:"type"`
	Price             float64 `json:"price"`
	AvailableQuantity int     `json:"available_quantity"`
}