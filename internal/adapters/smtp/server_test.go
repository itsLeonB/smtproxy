package smtp

import (
	"context"
	"testing"

	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	port := "2525"
	maxSize := int64(1024)
	authUsers := map[string]string{"user": "pass"}
	registry := provider.NewRegistry()

	server := NewServer(port, maxSize, authUsers, true, registry)

	assert.NotNil(t, server)
	assert.Equal(t, ":2525", server.addr)
	assert.NotNil(t, server.server)
	assert.Equal(t, ":2525", server.server.Addr)
	assert.Equal(t, "localhost", server.server.Domain)
	assert.Equal(t, maxSize, server.server.MaxMessageBytes)
	assert.True(t, server.server.AllowInsecureAuth)
}

func TestNewServer_NoAuth(t *testing.T) {
	port := "2525"
	maxSize := int64(1024)

	server := NewServer(port, maxSize, nil, false, nil)

	assert.NotNil(t, server)
	assert.Equal(t, ":2525", server.addr)
}

func TestServer_StartAndShutdown(t *testing.T) {
	server := NewServer("0", 1024, nil, false, nil) // Use port 0 for random available port

	err := server.Start()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}
