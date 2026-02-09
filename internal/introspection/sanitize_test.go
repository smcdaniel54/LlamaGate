package introspection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeConfig_redacts(t *testing.T) {
	m := map[string]interface{}{
		"api_key":    "secret123",
		"name":       "hello",
		"password":   "pw",
		"token":      "t",
		"nested": map[string]interface{}{
			"auth": "hidden",
			"ok":   "visible",
		},
	}
	out := SanitizeConfig(m)
	assert.Equal(t, redactedValue, out["api_key"])
	assert.Equal(t, "hello", out["name"])
	assert.Equal(t, redactedValue, out["password"])
	assert.Equal(t, redactedValue, out["token"])
	nested, ok := out["nested"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, redactedValue, nested["auth"])
	assert.Equal(t, "visible", nested["ok"])
}

func TestSanitizeConfig_nil(t *testing.T) {
	assert.Nil(t, SanitizeConfig(nil))
}

func Test_isSecretKey(t *testing.T) {
	assert.True(t, isSecretKey("api_key"))
	assert.True(t, isSecretKey("APIKey"))
	assert.True(t, isSecretKey("my_password"))
	assert.False(t, isSecretKey("name"))
	assert.False(t, isSecretKey("path"))
}
