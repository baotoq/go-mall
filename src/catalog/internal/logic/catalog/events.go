package catalog

import (
	"time"

	"github.com/google/uuid"
)

type ProductCreatedEvent struct {
	OccurredAt time.Time `json:"occurred_at"`
	ProductID  uuid.UUID `json:"id"`
}

func (e ProductCreatedEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type ProductUpdatedEvent struct {
	OccurredAt time.Time `json:"occurred_at"`
	ProductID  uuid.UUID `json:"id"`
}

func (e ProductUpdatedEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type ProductDeletedEvent struct {
	OccurredAt time.Time `json:"occurred_at"`
	ProductID  uuid.UUID `json:"id"`
}

func (e ProductDeletedEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type ProductStockIncreasedEvent struct {
	OccurredAt time.Time `json:"occurred_at"`
	ProductID  uuid.UUID `json:"id"`
	Quantity   int64     `json:"quantity"`
}

func (e ProductStockIncreasedEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type CategoryCreatedEvent struct {
	OccurredAt time.Time `json:"occurred_at"`
	CategoryID uuid.UUID `json:"id"`
}

func (e CategoryCreatedEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type ReservationCreatedEvent struct {
	OccurredAt    time.Time `json:"occurred_at"`
	ReservationID uuid.UUID `json:"id"`
	SessionID     string    `json:"session_id"`
}

func (e ReservationCreatedEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type ReservationConfirmedEvent struct {
	OccurredAt    time.Time `json:"occurred_at"`
	ReservationID uuid.UUID `json:"id"`
}

func (e ReservationConfirmedEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type ReservationCancelledEvent struct {
	OccurredAt    time.Time `json:"occurred_at"`
	ReservationID uuid.UUID `json:"id"`
}

func (e ReservationCancelledEvent) EventID() uuid.UUID { return uuid.Must(uuid.NewV7()) }
