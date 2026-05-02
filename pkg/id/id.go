// Package id provides UUIDv7 helpers for time-sortable ID generation.
package id

import (
	"fmt"

	"github.com/google/uuid"
)

// New generates a new UUIDv7. If RNG fails (extremely rare), panics with wrapped error.
func New() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Errorf("failed to generate UUIDv7: %w", err))
	}
	return id
}

// NewString returns a new UUIDv7 as a string.
func NewString() string {
	return New().String()
}

// MustParse parses a UUID string, panicking if invalid.
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// Parse parses a UUID string, returning an error if invalid.
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
