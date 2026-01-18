package provider

import (
	"context"
	"testing"

	"github.com/itsLeonB/smtproxy/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	provider := NewMockProvider("test-provider")
	
	err := registry.Register(provider)
	assert.NoError(t, err)
	
	// Should set as default since it's the first
	defaultProvider, err := registry.GetDefault()
	assert.NoError(t, err)
	assert.Equal(t, "test-provider", defaultProvider.Name())
}

func TestRegistry_RegisterEmptyName(t *testing.T) {
	registry := NewRegistry()
	provider := NewMockProvider("")
	
	err := registry.Register(provider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider name cannot be empty")
}

func TestRegistry_SetDefault(t *testing.T) {
	registry := NewRegistry()
	provider1 := NewMockProvider("provider1")
	provider2 := NewMockProvider("provider2")
	
	registry.Register(provider1)
	registry.Register(provider2)
	
	err := registry.SetDefault("provider2")
	assert.NoError(t, err)
	
	defaultProvider, err := registry.GetDefault()
	assert.NoError(t, err)
	assert.Equal(t, "provider2", defaultProvider.Name())
}

func TestRegistry_SetDefaultNotFound(t *testing.T) {
	registry := NewRegistry()
	
	err := registry.SetDefault("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestRegistry_GetProvider(t *testing.T) {
	registry := NewRegistry()
	provider := NewMockProvider("test-provider")
	
	registry.Register(provider)
	
	retrieved, err := registry.GetProvider("test-provider")
	assert.NoError(t, err)
	assert.Equal(t, "test-provider", retrieved.Name())
}

func TestRegistry_GetProviderNotFound(t *testing.T) {
	registry := NewRegistry()
	
	_, err := registry.GetProvider("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestRegistry_GetDefaultNoProviders(t *testing.T) {
	registry := NewRegistry()
	
	_, err := registry.GetDefault()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no default provider configured")
}

func TestRegistry_Send(t *testing.T) {
	registry := NewRegistry()
	provider := NewMockProvider("test-provider")
	registry.Register(provider)
	
	email := &entity.Email{}
	
	result, err := registry.Send(context.Background(), email, "")
	assert.NoError(t, err)
	assert.Equal(t, "test-provider", result.ProviderName)
	assert.NoError(t, result.Error)
}

func TestRegistry_SendSpecificProvider(t *testing.T) {
	registry := NewRegistry()
	provider1 := NewMockProvider("provider1")
	provider2 := NewMockProvider("provider2")
	
	registry.Register(provider1)
	registry.Register(provider2)
	
	email := &entity.Email{}
	
	result, err := registry.Send(context.Background(), email, "provider2")
	assert.NoError(t, err)
	assert.Equal(t, "provider2", result.ProviderName)
}

func TestRegistry_SendProviderNotFound(t *testing.T) {
	registry := NewRegistry()
	email := &entity.Email{}
	
	result, err := registry.Send(context.Background(), email, "nonexistent")
	assert.Error(t, err)
	assert.NotNil(t, result.Error)
}

func TestRegistry_ListProviders(t *testing.T) {
	registry := NewRegistry()
	provider1 := NewMockProvider("provider1")
	provider2 := NewMockProvider("provider2")
	
	registry.Register(provider1)
	registry.Register(provider2)
	
	providers := registry.ListProviders()
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, "provider1")
	assert.Contains(t, providers, "provider2")
}
