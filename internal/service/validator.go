package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var configKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+\.[a-zA-Z0-9_.-]+$`)

const maxConfigValueLength = 4096

func validateConfigs(configs map[string]string) error {
	if len(configs) == 0 {
		return fmt.Errorf("configs is required")
	}

	for key, value := range configs {
		if !configKeyPattern.MatchString(key) {
			return fmt.Errorf("invalid config key: %s", key)
		}

		if len(value) > maxConfigValueLength {
			return fmt.Errorf("config value too long for key: %s", key)
		}

		trimmed := strings.TrimSpace(value)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			var target any
			if err := json.Unmarshal([]byte(trimmed), &target); err != nil {
				return fmt.Errorf("invalid json config for key %s: %w", key, err)
			}
		}
	}

	return nil
}
