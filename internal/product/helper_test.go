package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIDFromString(t *testing.T) {
	t.Run("should parse valid ID successfully", func(t *testing.T) {
		id, err := ParseIDFromString("123")
		assert.NoError(t, err)
		assert.Equal(t, uint(123), id)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		_, err := ParseIDFromString("invalid")
		assert.Error(t, err)
	})

	t.Run("should parse negative ID", func(t *testing.T) {
		id, err := ParseIDFromString("-1")
		assert.NoError(t, err)
		assert.NotZero(t, id)
	})

	t.Run("should parse zero ID", func(t *testing.T) {
		id, err := ParseIDFromString("0")
		assert.NoError(t, err)
		assert.Equal(t, uint(0), id)
	})

	t.Run("should return error for empty string", func(t *testing.T) {
		_, err := ParseIDFromString("")
		assert.Error(t, err)
	})
}
