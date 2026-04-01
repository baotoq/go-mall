package event

import "github.com/google/uuid"

type Event interface {
	EventID() uuid.UUID
}
