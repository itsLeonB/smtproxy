package smtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientIdentity(t *testing.T) {
	identity := NewClientIdentity("testuser")
	
	assert.NotNil(t, identity)
	assert.Equal(t, "testuser", identity.Username)
	assert.True(t, identity.Authenticated)
}

func TestClientIdentity_IsAuthenticated(t *testing.T) {
	identity := &ClientIdentity{
		Username:      "testuser",
		Authenticated: true,
	}
	
	assert.True(t, identity.IsAuthenticated())
	
	identity.Authenticated = false
	assert.False(t, identity.IsAuthenticated())
}
