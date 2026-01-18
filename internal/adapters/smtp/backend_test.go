package smtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBackend(t *testing.T) {
	maxSize := int64(1024)
	users := map[string]string{"user": "pass"}
	authHandler := NewAuthHandler(users)
	
	backend := NewBackend(maxSize, authHandler, true)
	
	assert.NotNil(t, backend)
	assert.Equal(t, maxSize, backend.maxMessageSize)
	assert.Equal(t, authHandler, backend.authHandler)
	assert.True(t, backend.authEnabled)
}

func TestBackend_NewSession(t *testing.T) {
	backend := NewBackend(1024, nil, false)
	session, err := backend.NewSession(nil)
	
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.IsType(t, &Session{}, session)
}
