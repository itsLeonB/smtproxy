package config

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

func TestConfig_AuthUsers_EnvParsing(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected map[string]string
	}{
		{
			name:     "valid users",
			envValue: "user1:pass1,user2:pass2",
			expected: map[string]string{
				"user1": "pass1",
				"user2": "pass2",
			},
		},
		{
			name:     "single user",
			envValue: "admin:secret",
			expected: map[string]string{
				"admin": "secret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			_ = os.Setenv("AUTH_USERS", tt.envValue)
			defer func() {
				_ = os.Unsetenv("AUTH_USERS")
			}()

			var cfg Config
			err := envconfig.Process("", &cfg)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.AuthUsers)
		})
	}
}
