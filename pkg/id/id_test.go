package id_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gomall/pkg/id"
)

func TestNew_ReturnsValidV7(t *testing.T) {
	// Arrange & Act
	u := id.New()

	// Assert
	assert.Equal(t, uuid.Version(7), u.Version())
	assert.Equal(t, uuid.RFC4122, u.Variant())
}

func TestNewString_ParseRoundtrip(t *testing.T) {
	// Arrange
	s := id.NewString()

	// Act
	parsed, err := id.Parse(s)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, s, parsed.String())
}

func TestNew_TimeOrdering(t *testing.T) {
	// Arrange & Act — sleep well past UUIDv7 millisecond precision to avoid
	// same-ms collisions on fast hardware / loaded CI.
	const gap = 5 * time.Millisecond
	ids := make([]string, 10)
	for i := range ids {
		ids[i] = id.NewString()
		time.Sleep(gap)
	}

	// Assert
	for i := 1; i < len(ids); i++ {
		assert.Less(t, ids[i-1], ids[i],
			"UUIDv7 ids generated %s apart should be lexicographically ordered", gap)
	}
}

func TestMustParse_PanicsOnInvalid(t *testing.T) {
	// Arrange, Act & Assert
	assert.Panics(t, func() {
		id.MustParse("not-a-uuid")
	})
}

func TestParse_ErrorOnInvalid(t *testing.T) {
	// Arrange, Act & Assert
	_, err := id.Parse("not-a-uuid")
	assert.Error(t, err)
}
