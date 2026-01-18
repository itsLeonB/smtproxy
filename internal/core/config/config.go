package config

import (
	"github.com/itsLeonB/ungerr"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	LogLevel         string            `envconfig:"LOG_LEVEL" default:"info"`
	SMTPAddr         string            `envconfig:"SMTP_ADDR" default:":2525"`
	MaxSize          int64             `envconfig:"MAX_MESSAGE_SIZE" default:"10485760"` // 10MB
	AuthEnabled      bool              `envconfig:"AUTH_ENABLED" default:"true"`
	AuthUsers        map[string]string `envconfig:"AUTH_USERS" default:"user1:pass1,user2:pass2"`
	DefaultProvider  string            `envconfig:"DEFAULT_PROVIDER" default:""`
	EnabledProviders string            `envconfig:"ENABLED_PROVIDERS" default:""`
}

var Global *Config

func Load() error {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return ungerr.Wrap(err, "error loading env vars")
	}
	Global = &cfg
	return nil
}
