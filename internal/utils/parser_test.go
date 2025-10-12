package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIDFromString(t *testing.T) {
	t.Run("should parse valid positive ID", func(t *testing.T) {
		id, err := ParseIDFromString("123")
		assert.NoError(t, err)
		assert.Equal(t, uint(123), id)
	})

	t.Run("should parse large ID", func(t *testing.T) {
		id, err := ParseIDFromString("999999")
		assert.NoError(t, err)
		assert.Equal(t, uint(999999), id)
	})

	t.Run("should parse zero ID", func(t *testing.T) {
		id, err := ParseIDFromString("0")
		assert.NoError(t, err)
		assert.Equal(t, uint(0), id)
	})

	t.Run("should parse negative ID as uint wraparound", func(t *testing.T) {
		id, err := ParseIDFromString("-1")
		assert.NoError(t, err)
		assert.NotZero(t, id)
	})

	t.Run("should return error for invalid numeric format", func(t *testing.T) {
		_, err := ParseIDFromString("abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("should return error for empty string", func(t *testing.T) {
		_, err := ParseIDFromString("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("should return error for string with spaces", func(t *testing.T) {
		_, err := ParseIDFromString("12 34")
		assert.Error(t, err)
	})

	t.Run("should return error for alphanumeric string", func(t *testing.T) {
		_, err := ParseIDFromString("123abc")
		assert.Error(t, err)
	})

	t.Run("should return error for float string", func(t *testing.T) {
		_, err := ParseIDFromString("123.45")
		assert.Error(t, err)
	})

	t.Run("should return error for special characters", func(t *testing.T) {
		_, err := ParseIDFromString("@#$")
		assert.Error(t, err)
	})
}

func TestParseUserIDFromString(t *testing.T) {
	t.Run("should parse valid positive user ID", func(t *testing.T) {
		id, err := ParseUserIDFromString("456")
		assert.NoError(t, err)
		assert.Equal(t, uint(456), id)
	})

	t.Run("should parse large user ID", func(t *testing.T) {
		id, err := ParseUserIDFromString("4294967295")
		assert.NoError(t, err)
		assert.Equal(t, uint(4294967295), id)
	})

	t.Run("should parse zero user ID", func(t *testing.T) {
		id, err := ParseUserIDFromString("0")
		assert.NoError(t, err)
		assert.Equal(t, uint(0), id)
	})

	t.Run("should return error for negative user ID", func(t *testing.T) {
		_, err := ParseUserIDFromString("-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("should return error for invalid numeric format", func(t *testing.T) {
		_, err := ParseUserIDFromString("xyz")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("should return error for empty string", func(t *testing.T) {
		_, err := ParseUserIDFromString("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid syntax")
	})

	t.Run("should return error for string with spaces", func(t *testing.T) {
		_, err := ParseUserIDFromString("45 67")
		assert.Error(t, err)
	})

	t.Run("should return error for alphanumeric string", func(t *testing.T) {
		_, err := ParseUserIDFromString("456abc")
		assert.Error(t, err)
	})

	t.Run("should return error for float string", func(t *testing.T) {
		_, err := ParseUserIDFromString("456.78")
		assert.Error(t, err)
	})

	t.Run("should return error for special characters", func(t *testing.T) {
		_, err := ParseUserIDFromString("!@#")
		assert.Error(t, err)
	})

	t.Run("should return error for number exceeding uint32 max", func(t *testing.T) {
		_, err := ParseUserIDFromString("4294967296")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "value out of range")
	})
}
