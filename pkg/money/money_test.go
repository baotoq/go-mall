package money_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"gomall/pkg/money"
)

func TestNew_NormalizesCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"usd", "USD"},
		{" usd ", "USD"},
		{"", "USD"},
		{"EUR", "EUR"},
	}
	for _, tc := range tests {
		// Arrange + Act
		m := money.New(100, tc.input)
		// Assert
		assert.Equal(t, tc.expected, m.Currency, "input: %q", tc.input)
	}
}

func TestAdd_SameCurrency(t *testing.T) {
	// Arrange
	a := money.New(100, "USD")
	b := money.New(50, "USD")

	// Act
	result, err := a.Add(b)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, money.New(150, "USD"), result)
}

func TestAdd_CurrencyMismatch(t *testing.T) {
	// Arrange
	a := money.New(100, "USD")
	b := money.New(50, "EUR")

	// Act
	_, err := a.Add(b)

	// Assert
	assert.ErrorIs(t, err, money.ErrCurrencyMismatch)
}

func TestAdd_Overflow(t *testing.T) {
	t.Run("positive overflow", func(t *testing.T) {
		// Arrange
		a := money.New(math.MaxInt64, "USD")
		b := money.New(1, "USD")

		// Act
		_, err := a.Add(b)

		// Assert
		assert.ErrorIs(t, err, money.ErrOverflow)
	})

	t.Run("negative overflow", func(t *testing.T) {
		// Arrange
		a := money.New(math.MinInt64, "USD")
		b := money.New(-1, "USD")

		// Act
		_, err := a.Add(b)

		// Assert
		assert.ErrorIs(t, err, money.ErrOverflow)
	})
}

func TestSub_SameCurrency(t *testing.T) {
	// Arrange
	a := money.New(50, "USD")
	b := money.New(100, "USD")

	// Act
	result, err := a.Sub(b)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, money.New(-50, "USD"), result)
	assert.True(t, result.IsNegative())
}

func TestSub_Overflow(t *testing.T) {
	// Arrange — MinInt64 - 1 overflows.
	a := money.New(math.MinInt64, "USD")
	b := money.New(1, "USD")

	// Act
	_, err := a.Sub(b)

	// Assert
	assert.ErrorIs(t, err, money.ErrOverflow)
}

func TestMul_Basic(t *testing.T) {
	// Arrange
	m := money.New(100, "USD")

	// Act + Assert
	got, err := m.Mul(3)
	assert.NoError(t, err)
	assert.Equal(t, money.New(300, "USD"), got)

	zero, err := m.Mul(0)
	assert.NoError(t, err)
	assert.True(t, zero.IsZero())

	neg, err := m.Mul(-1)
	assert.NoError(t, err)
	assert.Equal(t, money.New(-100, "USD"), neg)
}

func TestMul_Overflow(t *testing.T) {
	t.Run("MaxInt64 * 2", func(t *testing.T) {
		// Arrange + Act
		_, err := money.New(math.MaxInt64, "USD").Mul(2)
		// Assert
		assert.ErrorIs(t, err, money.ErrOverflow)
	})

	t.Run("MinInt64 * -1", func(t *testing.T) {
		// Arrange + Act — would also panic without explicit guard.
		_, err := money.New(math.MinInt64, "USD").Mul(-1)
		// Assert
		assert.ErrorIs(t, err, money.ErrOverflow)
	})
}

func TestNeg(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		// Arrange + Act
		got, err := money.New(100, "USD").Neg()
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, money.New(-100, "USD"), got)
	})

	t.Run("MinInt64 overflows", func(t *testing.T) {
		// Arrange + Act
		_, err := money.New(math.MinInt64, "USD").Neg()
		// Assert
		assert.ErrorIs(t, err, money.ErrOverflow)
	})
}

func TestCmp(t *testing.T) {
	a := money.New(100, "USD")

	t.Run("less", func(t *testing.T) {
		c, err := a.Cmp(money.New(200, "USD"))
		assert.NoError(t, err)
		assert.Equal(t, -1, c)
	})
	t.Run("equal", func(t *testing.T) {
		c, err := a.Cmp(money.New(100, "USD"))
		assert.NoError(t, err)
		assert.Equal(t, 0, c)
	})
	t.Run("greater", func(t *testing.T) {
		c, err := a.Cmp(money.New(50, "USD"))
		assert.NoError(t, err)
		assert.Equal(t, 1, c)
	})
	t.Run("currency mismatch", func(t *testing.T) {
		_, err := a.Cmp(money.New(100, "EUR"))
		assert.ErrorIs(t, err, money.ErrCurrencyMismatch)
	})
}

func TestEqual(t *testing.T) {
	// Arrange
	a := money.New(100, "USD")

	// Assert
	assert.True(t, a.Equal(money.New(100, "USD")))
	assert.False(t, a.Equal(money.New(100, "EUR")))
	assert.False(t, a.Equal(money.New(200, "USD")))
}

func TestSum_Empty(t *testing.T) {
	// Act
	_, err := money.Sum(nil)

	// Assert
	assert.ErrorIs(t, err, money.ErrEmptySum)
}

func TestSum_Mixed(t *testing.T) {
	t.Run("three USD items", func(t *testing.T) {
		// Arrange
		items := []money.Money{
			money.New(100, "USD"),
			money.New(200, "USD"),
			money.New(300, "USD"),
		}

		// Act
		result, err := money.Sum(items)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, money.New(600, "USD"), result)
	})

	t.Run("mixing EUR returns mismatch", func(t *testing.T) {
		// Arrange
		items := []money.Money{
			money.New(100, "USD"),
			money.New(200, "USD"),
			money.New(50, "EUR"),
		}

		// Act
		_, err := money.Sum(items)

		// Assert
		assert.ErrorIs(t, err, money.ErrCurrencyMismatch)
	})

	t.Run("overflow propagates", func(t *testing.T) {
		// Arrange
		items := []money.Money{
			money.New(math.MaxInt64, "USD"),
			money.New(1, "USD"),
		}

		// Act
		_, err := money.Sum(items)

		// Assert
		assert.ErrorIs(t, err, money.ErrOverflow)
	})
}

func TestString_FormatCases(t *testing.T) {
	tests := []struct {
		cents    int64
		expected string
	}{
		{0, "USD 0.00"},
		{123, "USD 1.23"},
		{100, "USD 1.00"},
		{99, "USD 0.99"},
		{-99, "USD -0.99"},
		{-1234, "USD -12.34"},
		{math.MaxInt64, "USD 92233720368547758.07"},
		{math.MinInt64, "USD -92233720368547758.08"},
	}
	for _, tc := range tests {
		// Arrange + Act
		m := money.New(tc.cents, "USD")
		// Assert
		assert.Equal(t, tc.expected, m.String(), "cents=%d", tc.cents)
	}
}

func TestJSON_Roundtrip(t *testing.T) {
	t.Run("roundtrip preserves value", func(t *testing.T) {
		// Arrange
		original := money.New(1234, "USD")

		// Act
		data, err := json.Marshal(original)
		assert.NoError(t, err)

		var result money.Money
		err = json.Unmarshal(data, &result)

		// Assert
		assert.NoError(t, err)
		assert.True(t, original.Equal(result))
	})

	t.Run("unmarshal normalizes currency case", func(t *testing.T) {
		// Arrange — lowercase incoming JSON should be normalized to uppercase
		// so values round-trip cleanly with money.New("USD") on Add/Equal.
		raw := `{"cents":100,"currency":"usd"}`

		// Act
		var m money.Money
		err := json.Unmarshal([]byte(raw), &m)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "USD", m.Currency)
		assert.Equal(t, int64(100), m.Cents)
		// Unmarshalled value adds cleanly to New-constructed value.
		sum, err := m.Add(money.New(50, "USD"))
		assert.NoError(t, err)
		assert.Equal(t, money.New(150, "USD"), sum)
	})

	t.Run("unmarshal empty currency defaults to USD", func(t *testing.T) {
		// Arrange
		raw := `{"cents":100,"currency":""}`

		// Act
		var m money.Money
		err := json.Unmarshal([]byte(raw), &m)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "USD", m.Currency)
	})
}
