package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	LogLevel         string            `envconfig:"LOG_LEVEL" default:"info"`
	SMTPPort         string            `envconfig:"SMTP_PORT" default:"2525"`
	MaxSize          int64             `envconfig:"MAX_MESSAGE_SIZE" default:"10485760"` // 10MB
	AuthEnabled      bool              `envconfig:"AUTH_ENABLED" default:"true"`
	AuthUsers        map[string]string `envconfig:"AUTH_USERS" default:"user1:pass1,user2:pass2"`
	DefaultProvider  string            `envconfig:"DEFAULT_PROVIDER" default:"brevo"`
	EnabledProviders string            `envconfig:"ENABLED_PROVIDERS" default:"brevo"`

	// Brevo configuration
	BrevoAPIKey  string        `envconfig:"BREVO_API_KEY"`
	BrevoBaseURL string        `envconfig:"BREVO_BASE_URL" default:"https://api.brevo.com/v3"`
	BrevoTimeout time.Duration `envconfig:"BREVO_TIMEOUT" default:"30s"`
}

var Global *Config

func Load() error {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return err
	}
	Global = &cfg
	return nil
}
