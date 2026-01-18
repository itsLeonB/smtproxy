package provider

import (
	"context"
	"errors"
	"sync"

	"github.com/itsLeonB/smtproxy/internal/domain/entity"
)

// Registry manages registered providers and routing
type Registry struct {
	mu              sync.RWMutex
	providers       map[string]Provider
	defaultProvider string
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := provider.Name()
	if name == "" {
		return errors.New("provider name cannot be empty")
	}
	
	r.providers[name] = provider
	
	// Set as default if it's the first provider
	if r.defaultProvider == "" {
		r.defaultProvider = name
	}
	
	return nil
}

// SetDefault sets the default provider
func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.providers[name]; !exists {
		return errors.New("provider not found: " + name)
	}
	
	r.defaultProvider = name
	return nil
}

// GetProvider returns a provider by name
func (r *Registry) GetProvider(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	provider, exists := r.providers[name]
	if !exists {
		return nil, errors.New("provider not found: " + name)
	}
	
	return provider, nil
}

// GetDefault returns the default provider
func (r *Registry) GetDefault() (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.defaultProvider == "" {
		return nil, errors.New("no default provider configured")
	}
	
	return r.providers[r.defaultProvider], nil
}

// Send routes email to the specified provider or default
func (r *Registry) Send(ctx context.Context, email *entity.Email, providerName string) (*SendResult, error) {
	var provider Provider
	var err error
	
	if providerName != "" {
		provider, err = r.GetProvider(providerName)
		if err != nil {
			return &SendResult{Error: err}, err
		}
	} else {
		provider, err = r.GetDefault()
		if err != nil {
			return &SendResult{Error: err}, err
		}
	}
	
	err = provider.Send(ctx, email)
	return &SendResult{
		ProviderName: provider.Name(),
		Error:        err,
	}, err
}

// ListProviders returns all registered provider names
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
