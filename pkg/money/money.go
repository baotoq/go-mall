// Package money provides a simple Money type for carrying cents+currency together.
package money

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

// ErrCurrencyMismatch is returned when an operation is attempted on Money values with different currencies.
var ErrCurrencyMismatch = errors.New("currency mismatch")

// ErrEmptySum is returned by Sum when called with an empty slice.
var ErrEmptySum = errors.New("sum of empty slice")

// ErrOverflow is returned when an arithmetic operation would overflow int64.
var ErrOverflow = errors.New("money overflow")

// Money holds an amount in the smallest currency unit (cents) and a currency code.
//
// Currency is normalized (uppercased, trimmed) by New, Zero, and UnmarshalJSON.
// Direct struct construction (Money{Cents: ..., Currency: "usd"}) bypasses
// normalization and may break Add/Sub equality checks; prefer the constructors.
type Money struct {
	Cents    int64  `json:"cents"`
	Currency string `json:"currency"`
}

// normalizeCurrency uppercases and trims the currency string, defaulting to "USD" if empty.
//
// Defaulting empty to USD is intentional ergonomics for ent zero-values and
// callers that omit the field; pass an explicit non-empty code to opt out.
func normalizeCurrency(currency string) string {
	c := strings.TrimSpace(strings.ToUpper(currency))
	if c == "" {
		return "USD"
	}
	return c
}

// New returns a Money value with the given cents and a normalized currency code.
func New(cents int64, currency string) Money {
	return Money{Cents: cents, Currency: normalizeCurrency(currency)}
}

// Zero returns a Money value with zero cents and a normalized currency code.
func Zero(currency string) Money {
	return New(0, currency)
}

// Add returns the sum of m and o. Returns ErrCurrencyMismatch if currencies
// differ, or ErrOverflow if the result would overflow int64.
func (m Money) Add(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	sum := m.Cents + o.Cents
	// Overflow iff operands share a sign and the result's sign flips.
	if (m.Cents > 0 && o.Cents > 0 && sum < 0) ||
		(m.Cents < 0 && o.Cents < 0 && sum >= 0) {
		return Money{}, ErrOverflow
	}
	return Money{Cents: sum, Currency: m.Currency}, nil
}

// Sub returns the difference of m and o. Returns ErrCurrencyMismatch if
// currencies differ, or ErrOverflow if the result would overflow int64.
func (m Money) Sub(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	diff := m.Cents - o.Cents
	// Overflow iff operands have opposite signs and result's sign disagrees with m.
	if (m.Cents >= 0 && o.Cents < 0 && diff < 0) ||
		(m.Cents < 0 && o.Cents > 0 && diff >= 0) {
		return Money{}, ErrOverflow
	}
	return Money{Cents: diff, Currency: m.Currency}, nil
}

// Mul returns m with its cents multiplied by n. Returns ErrOverflow if the
// result would overflow int64.
func (m Money) Mul(n int64) (Money, error) {
	if m.Cents == 0 || n == 0 {
		return Money{Cents: 0, Currency: m.Currency}, nil
	}
	// MinInt64 / -1 panics with integer-overflow; reject before division check.
	if (m.Cents == math.MinInt64 && n == -1) || (n == math.MinInt64 && m.Cents == -1) {
		return Money{}, ErrOverflow
	}
	prod := m.Cents * n
	if prod/n != m.Cents {
		return Money{}, ErrOverflow
	}
	return Money{Cents: prod, Currency: m.Currency}, nil
}

// Neg returns -m. Returns ErrOverflow if m.Cents == math.MinInt64.
func (m Money) Neg() (Money, error) {
	if m.Cents == math.MinInt64 {
		return Money{}, ErrOverflow
	}
	return Money{Cents: -m.Cents, Currency: m.Currency}, nil
}

// Cmp compares m and o, returning -1, 0, or 1. Returns ErrCurrencyMismatch if
// currencies differ.
func (m Money) Cmp(o Money) (int, error) {
	if m.Currency != o.Currency {
		return 0, ErrCurrencyMismatch
	}
	switch {
	case m.Cents < o.Cents:
		return -1, nil
	case m.Cents > o.Cents:
		return 1, nil
	default:
		return 0, nil
	}
}

// Equal reports whether m and o have the same cents and currency.
func (m Money) Equal(o Money) bool {
	return m.Cents == o.Cents && m.Currency == o.Currency
}

// IsZero reports whether m has zero cents.
func (m Money) IsZero() bool {
	return m.Cents == 0
}

// IsNegative reports whether m has negative cents.
func (m Money) IsNegative() bool {
	return m.Cents < 0
}

// String returns a human-readable representation like "USD 12.34" or "USD -0.99".
// This is a debug representation only; it is not locale-aware. Handles
// math.MinInt64 correctly via uint64 negation.
func (m Money) String() string {
	abs := uint64(m.Cents)
	if m.Cents < 0 {
		abs = -abs
	}
	major := abs / 100
	minor := abs % 100
	if m.Cents < 0 {
		return fmt.Sprintf("%s -%d.%02d", m.Currency, major, minor)
	}
	return fmt.Sprintf("%s %d.%02d", m.Currency, major, minor)
}

// UnmarshalJSON decodes Money and normalizes the currency code so that JSON
// inputs round-trip correctly through Add/Sub/Equal alongside Money values
// constructed via New.
func (m *Money) UnmarshalJSON(data []byte) error {
	var raw struct {
		Cents    int64  `json:"cents"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Cents = raw.Cents
	m.Currency = normalizeCurrency(raw.Currency)
	return nil
}

// Sum returns the total of all Money values in items.
// Returns ErrEmptySum if items is empty.
// Returns ErrCurrencyMismatch if any item has a different currency than the first.
// Returns ErrOverflow if the running total would overflow int64.
func Sum(items []Money) (Money, error) {
	if len(items) == 0 {
		return Money{}, ErrEmptySum
	}
	acc := items[0]
	for i, item := range items[1:] {
		next, err := acc.Add(item)
		if err != nil {
			return Money{}, fmt.Errorf("sum: item %d: %w", i+1, err)
		}
		acc = next
	}
	return acc, nil
}
