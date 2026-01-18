package smtp

import (
	"github.com/emersion/go-smtp"
	"github.com/itsLeonB/smtproxy/internal/domain/service/parser"
)

// Backend implements smtp.Backend interface
type Backend struct {
	maxMessageSize int64
	authHandler    *AuthHandler
	authEnabled    bool
}

// NewBackend creates a new SMTP backend
func NewBackend(maxMessageSize int64, authHandler *AuthHandler, authEnabled bool) *Backend {
	return &Backend{
		maxMessageSize: maxMessageSize,
		authHandler:    authHandler,
		authEnabled:    authEnabled,
	}
}

// NewSession creates a new SMTP session
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		maxMessageSize: b.maxMessageSize,
		authHandler:    b.authHandler,
		authEnabled:    b.authEnabled,
		parser:         parser.New(b.maxMessageSize),
	}, nil
}
