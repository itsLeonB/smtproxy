package dispatcher

import (
	"context"
	"fmt"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/smtproxy/internal/domain/entity"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
)

// Dispatcher handles the core email dispatch flow
type Dispatcher struct {
	registry *provider.Registry
	logger   ezutil.Logger
}

// NewDispatcher creates a new email dispatcher
func NewDispatcher(registry *provider.Registry, logger ezutil.Logger) *Dispatcher {
	return &Dispatcher{
		registry: registry,
		logger:   logger,
	}
}

// Dispatch sends an email through the provider system
func (d *Dispatcher) Dispatch(ctx context.Context, email *entity.Email, providerName string) error {
	// Log send attempt
	if providerName != "" {
		d.logger.Infof("dispatching email to provider: %s", providerName)
	} else {
		d.logger.Info("dispatching email to default provider")
	}

	// Send via registry
	result, err := d.registry.Send(ctx, email, providerName)
	
	// Log result
	if err != nil {
		d.logger.Errorf("email dispatch failed - provider: %s, error: %v", result.ProviderName, err)
		return d.translateError(err)
	}
	
	d.logger.Infof("email dispatched successfully - provider: %s", result.ProviderName)
	return nil
}

// translateError converts provider errors to SMTP errors
func (d *Dispatcher) translateError(err error) error {
	if err == nil {
		return nil
	}

	// Map common provider errors to SMTP-friendly messages
	errMsg := err.Error()
	
	switch {
	case contains(errMsg, "authentication", "unauthorized", "invalid key", "forbidden"):
		return fmt.Errorf("550 Authentication failed")
	case contains(errMsg, "rate limit", "quota", "throttle"):
		return fmt.Errorf("451 Rate limit exceeded, try again later")
	case contains(errMsg, "invalid email", "invalid recipient", "bad address"):
		return fmt.Errorf("550 Invalid recipient address")
	case contains(errMsg, "timeout", "deadline"):
		return fmt.Errorf("451 Timeout occurred, try again later")
	case contains(errMsg, "service unavailable", "maintenance"):
		return fmt.Errorf("451 Service temporarily unavailable")
	default:
		return fmt.Errorf("451 Temporary failure: %s", errMsg)
	}
}

// contains checks if any of the keywords exist in the error message (case-insensitive)
func contains(msg string, keywords ...string) bool {
	msgLower := toLower(msg)
	for _, keyword := range keywords {
		if containsSubstring(msgLower, toLower(keyword)) {
			return true
		}
	}
	return false
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}

// containsSubstring checks if substring exists in string
func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
