package smtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	addr := ":2525"
	maxSize := int64(1024)
	
	server := NewServer(addr, maxSize)
	
	assert.NotNil(t, server)
	assert.Equal(t, addr, server.addr)
	assert.NotNil(t, server.server)
	assert.Equal(t, addr, server.server.Addr)
	assert.Equal(t, "localhost", server.server.Domain)
	assert.Equal(t, maxSize, server.server.MaxMessageBytes)
	assert.True(t, server.server.AllowInsecureAuth)
}

func TestServer_StartAndShutdown(t *testing.T) {
	server := NewServer(":0", 1024) // Use port 0 for random available port
	
	err := server.Start()
	assert.NoError(t, err)
	
	err = server.Shutdown(nil)
	assert.NoError(t, err)
}
