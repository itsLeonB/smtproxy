package smtp

import (
	"github.com/emersion/go-smtp"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/smtproxy/internal/domain/service/dispatcher"
	"github.com/itsLeonB/smtproxy/internal/domain/service/parser"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
)

// Backend implements smtp.Backend interface
type Backend struct {
	maxMessageSize int64
	authHandler    *AuthHandler
	authEnabled    bool
	dispatcher     *dispatcher.Dispatcher
}

// NewBackend creates a new SMTP backend
func NewBackend(maxMessageSize int64, authHandler *AuthHandler, authEnabled bool, registry *provider.Registry, logger ezutil.Logger) *Backend {
	var disp *dispatcher.Dispatcher
	if registry != nil {
		disp = dispatcher.NewDispatcher(registry, logger)
	}
	
	return &Backend{
		maxMessageSize: maxMessageSize,
		authHandler:    authHandler,
		authEnabled:    authEnabled,
		dispatcher:     disp,
	}
}

// NewSession creates a new SMTP session
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		maxMessageSize: b.maxMessageSize,
		authHandler:    b.authHandler,
		authEnabled:    b.authEnabled,
		parser:         parser.New(b.maxMessageSize),
		dispatcher:     b.dispatcher,
	}, nil
}

// AuthPlain implements SMTP AUTH PLAIN for the backend
func (b *Backend) AuthPlain(conn *smtp.Conn, username, password string) (smtp.Session, error) {
	session := &Session{
		maxMessageSize: b.maxMessageSize,
		authHandler:    b.authHandler,
		authEnabled:    b.authEnabled,
		parser:         parser.New(b.maxMessageSize),
		dispatcher:     b.dispatcher,
	}
	
	if err := session.AuthPlain(username, password); err != nil {
		return nil, err
	}
	
	return session, nil
}

// AuthLogin implements SMTP AUTH LOGIN for the backend
func (b *Backend) AuthLogin(conn *smtp.Conn, username, password string) (smtp.Session, error) {
	session := &Session{
		maxMessageSize: b.maxMessageSize,
		authHandler:    b.authHandler,
		authEnabled:    b.authEnabled,
		parser:         parser.New(b.maxMessageSize),
		dispatcher:     b.dispatcher,
	}
	
	if err := session.AuthLogin(username, password); err != nil {
		return nil, err
	}
	
	return session, nil
}
