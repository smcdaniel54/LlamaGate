package introspection

import (
	"encoding/json"
	"strings"
)

// SecretKeys are substrings (case-insensitive) in key names that trigger redaction.
var SecretKeys = []string{
	"api_key", "apikey", "token", "password", "secret", "auth",
	"authorization", "credential", "private",
}

const redactedValue = "<redacted>"

// SanitizeConfig recursively redacts map keys that look like secrets.
// Keys matching (case-insensitive) secret patterns have their value replaced with redactedValue.
func SanitizeConfig(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		if isSecretKey(k) {
			out[k] = redactedValue
			continue
		}
		out[k] = sanitizeValue(v)
	}
	return out
}

func isSecretKey(key string) bool {
	lower := strings.ToLower(key)
	for _, s := range SecretKeys {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

func sanitizeValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		return SanitizeConfig(x)
	case []interface{}:
		arr := make([]interface{}, len(x))
		for i, item := range x {
			arr[i] = sanitizeValue(item)
		}
		return arr
	default:
		return v
	}
}

// ConfigToMap converts a struct to a map for sanitization (using JSON round-trip).
func ConfigToMap(cfg interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
