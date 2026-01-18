package smtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBackend(t *testing.T) {
	maxSize := int64(1024)
	backend := NewBackend(maxSize)
	
	assert.NotNil(t, backend)
	assert.Equal(t, maxSize, backend.maxMessageSize)
}

func TestBackend_NewSession(t *testing.T) {
	backend := NewBackend(1024)
	session, err := backend.NewSession(nil)
	
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.IsType(t, &Session{}, session)
}
