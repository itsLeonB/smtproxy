package smtp

import (
	"github.com/emersion/go-smtp"
)

// Backend implements smtp.Backend interface
type Backend struct {
	maxMessageSize int64
}

// NewBackend creates a new SMTP backend
func NewBackend(maxMessageSize int64) *Backend {
	return &Backend{
		maxMessageSize: maxMessageSize,
	}
}

// NewSession creates a new SMTP session
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		maxMessageSize: b.maxMessageSize,
	}, nil
}
