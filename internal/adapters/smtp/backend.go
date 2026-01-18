package smtp

import (
	"github.com/emersion/go-smtp"
	"github.com/itsLeonB/smtproxy/internal/domain/service/parser"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
)

// Backend implements smtp.Backend interface
type Backend struct {
	maxMessageSize int64
	authHandler    *AuthHandler
	authEnabled    bool
	registry       *provider.Registry
}

// NewBackend creates a new SMTP backend
func NewBackend(maxMessageSize int64, authHandler *AuthHandler, authEnabled bool, registry *provider.Registry) *Backend {
	return &Backend{
		maxMessageSize: maxMessageSize,
		authHandler:    authHandler,
		authEnabled:    authEnabled,
		registry:       registry,
	}
}

// NewSession creates a new SMTP session
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		maxMessageSize: b.maxMessageSize,
		authHandler:    b.authHandler,
		authEnabled:    b.authEnabled,
		parser:         parser.New(b.maxMessageSize),
		registry:       b.registry,
	}, nil
}
