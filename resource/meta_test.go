package resource

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalStringMeta(t *testing.T) {
	raw := json.RawMessage(`{
		"monthly_price": "1.5",
		"setup_price": 0,
		"unpriced": null,
		"large": 9007199254740993,
		"enabled": true,
		"nested": {"a": 1}
	}`)

	meta, err := unmarshalStringMeta(raw)
	require.NoError(t, err)

	assert.Equal(t, "1.5", meta["monthly_price"])
	assert.Equal(t, "0", meta["setup_price"])
	assert.Equal(t, "", meta["unpriced"])
	// Kept verbatim: a float64 round trip would return 9007199254740992.
	assert.Equal(t, "9007199254740993", meta["large"])
	assert.Equal(t, "true", meta["enabled"])
	assert.Equal(t, `{"a": 1}`, meta["nested"])
}

func TestUnmarshalStringMetaRejectsNonObject(t *testing.T) {
	_, err := unmarshalStringMeta(json.RawMessage(`["not", "an", "object"]`))
	assert.Error(t, err)
}
