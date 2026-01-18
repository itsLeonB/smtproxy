package smtp

import (
	"context"
	"testing"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	addr := ":2525"
	maxSize := int64(1024)
	authUsers := map[string]string{"user": "pass"}
	registry := provider.NewRegistry()
	logger := ezutil.NewSimpleLogger("test", false, 1)

	server := NewServer(addr, maxSize, authUsers, true, registry, logger)

	assert.NotNil(t, server)
	assert.Equal(t, addr, server.addr)
	assert.NotNil(t, server.server)
	assert.Equal(t, addr, server.server.Addr)
	assert.Equal(t, "localhost", server.server.Domain)
	assert.Equal(t, maxSize, server.server.MaxMessageBytes)
	assert.True(t, server.server.AllowInsecureAuth)
}

func TestNewServer_NoAuth(t *testing.T) {
	addr := ":2525"
	maxSize := int64(1024)

	server := NewServer(addr, maxSize, nil, false, nil, nil)

	assert.NotNil(t, server)
	assert.Equal(t, addr, server.addr)
}

func TestServer_StartAndShutdown(t *testing.T) {
	server := NewServer(":0", 1024, nil, false, nil, nil) // Use port 0 for random available port

	err := server.Start()
	assert.NoError(t, err)

	err = server.Shutdown(context.Background())
	assert.NoError(t, err)
}
