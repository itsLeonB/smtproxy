package smtp

import (
	"context"
	"fmt"
	"net"

	"github.com/emersion/go-smtp"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
)

// Server wraps the SMTP server
type Server struct {
	server *smtp.Server
	addr   string
}

// NewServer creates a new SMTP server
func NewServer(addr string, maxMessageSize int64, authUsers map[string]string, authEnabled bool, registry *provider.Registry, logger ezutil.Logger) *Server {
	var authHandler *AuthHandler
	if authEnabled && len(authUsers) > 0 {
		authHandler = NewAuthHandler(authUsers)
	}
	
	backend := NewBackend(maxMessageSize, authHandler, authEnabled, registry, logger)
	
	s := smtp.NewServer(backend)
	s.Addr = addr
	s.Domain = "localhost"
	s.MaxMessageBytes = maxMessageSize
	s.AllowInsecureAuth = true

	return &Server{
		server: s,
		addr:   addr,
	}
}

// Start starts the SMTP server
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}

	go func() {
		if err := s.server.Serve(ln); err != nil {
			fmt.Printf("SMTP server error: %v\n", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the SMTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Close()
}
