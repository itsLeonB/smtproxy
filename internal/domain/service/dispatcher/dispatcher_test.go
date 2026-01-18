package dispatcher

import (
	"context"
	"errors"
	"testing"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/smtproxy/internal/domain/entity"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Dispatch_Success(t *testing.T) {
	registry := provider.NewRegistry()
	mockProvider := provider.NewMockProvider("test-provider")
	registry.Register(mockProvider)
	
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(registry, logger)
	
	email := &entity.Email{}
	
	err := dispatcher.Dispatch(context.Background(), email, "")
	assert.NoError(t, err)
}

func TestDispatcher_Dispatch_SpecificProvider(t *testing.T) {
	registry := provider.NewRegistry()
	mockProvider := provider.NewMockProvider("specific-provider")
	registry.Register(mockProvider)
	
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(registry, logger)
	
	email := &entity.Email{}
	
	err := dispatcher.Dispatch(context.Background(), email, "specific-provider")
	assert.NoError(t, err)
}

func TestDispatcher_Dispatch_ProviderError(t *testing.T) {
	registry := provider.NewRegistry()
	mockProvider := provider.NewMockProvider("failing-provider")
	mockProvider.SetSendError(errors.New("provider error"))
	registry.Register(mockProvider)
	
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(registry, logger)
	
	email := &entity.Email{}
	
	err := dispatcher.Dispatch(context.Background(), email, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "451 Temporary failure")
}

func TestDispatcher_TranslateError_Authentication(t *testing.T) {
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(nil, logger)
	
	err := errors.New("authentication failed")
	translated := dispatcher.translateError(err)
	
	assert.Contains(t, translated.Error(), "550 Authentication failed")
}

func TestDispatcher_TranslateError_RateLimit(t *testing.T) {
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(nil, logger)
	
	err := errors.New("rate limit exceeded")
	translated := dispatcher.translateError(err)
	
	assert.Contains(t, translated.Error(), "451 Rate limit exceeded")
}

func TestDispatcher_TranslateError_InvalidEmail(t *testing.T) {
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(nil, logger)
	
	err := errors.New("invalid email address")
	translated := dispatcher.translateError(err)
	
	assert.Contains(t, translated.Error(), "550 Invalid recipient")
}

func TestDispatcher_TranslateError_Timeout(t *testing.T) {
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(nil, logger)
	
	err := errors.New("request timeout")
	translated := dispatcher.translateError(err)
	
	assert.Contains(t, translated.Error(), "451 Timeout occurred")
}

func TestDispatcher_TranslateError_ServiceUnavailable(t *testing.T) {
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(nil, logger)
	
	err := errors.New("service unavailable")
	translated := dispatcher.translateError(err)
	
	assert.Contains(t, translated.Error(), "451 Service temporarily unavailable")
}

func TestDispatcher_TranslateError_Generic(t *testing.T) {
	logger := ezutil.NewSimpleLogger("test", false, 1)
	dispatcher := NewDispatcher(nil, logger)
	
	err := errors.New("unknown error")
	translated := dispatcher.translateError(err)
	
	assert.Contains(t, translated.Error(), "451 Temporary failure")
}
