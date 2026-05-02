package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsNotFound_nilError_returnsFalse(t *testing.T) {
	assert.False(t, isNotFound(nil))
}

func TestIsNotFound_grpcNotFoundCode_returnsTrue(t *testing.T) {
	// Arrange
	err := status.Error(codes.NotFound, "instance not found")
	// Act + Assert
	assert.True(t, isNotFound(err))
}

func TestIsNotFound_wrappedGrpcNotFound_returnsTrue(t *testing.T) {
	// Arrange
	inner := status.Error(codes.NotFound, "instance not found")
	wrapped := fmt.Errorf("outer: %w", inner)
	// Act + Assert
	assert.True(t, isNotFound(wrapped))
}

func TestIsNotFound_grpcOtherCode_returnsFalse(t *testing.T) {
	// Arrange
	err := status.Error(codes.Internal, "internal error")
	// Act + Assert
	assert.False(t, isNotFound(err))
}

func TestIsNotFound_plainError_returnsFalse(t *testing.T) {
	// Arrange: plain error containing "not found" text but NOT a gRPC NotFound code
	err := errors.New("something not found")
	// Act + Assert: after the fix, string matching is replaced by gRPC code check
	assert.False(t, isNotFound(err))
}
