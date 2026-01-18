package smtp

import (
	"testing"

	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
	"github.com/stretchr/testify/assert"
)

func TestNewBackend(t *testing.T) {
	maxSize := int64(1024)
	users := map[string]string{"user": "pass"}
	authHandler := NewAuthHandler(users)
	registry := provider.NewRegistry()

	backend := NewBackend(maxSize, authHandler, true, registry)

	assert.NotNil(t, backend)
	assert.Equal(t, maxSize, backend.maxMessageSize)
	assert.Equal(t, authHandler, backend.authHandler)
	assert.True(t, backend.authEnabled)
	assert.NotNil(t, backend.dispatcher)
}

func TestBackend_NewSession(t *testing.T) {
	backend := NewBackend(1024, nil, false, nil)
	session, err := backend.NewSession(nil)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.IsType(t, &Session{}, session)
}
