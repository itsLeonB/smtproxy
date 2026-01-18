package smtp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/emersion/go-smtp"
	"github.com/itsLeonB/smtproxy/internal/adapters/providers/brevo"
	"github.com/itsLeonB/smtproxy/internal/core/config"
	"github.com/itsLeonB/smtproxy/internal/core/logger"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
)

// Server wraps the SMTP server
type Server struct {
	server *smtp.Server
	addr   string
}

func Setup() (*Server, error) {
	authUsers := config.Global.AuthUsers

	// Initialize provider registry
	registry := provider.NewRegistry()

	// Register Brevo provider if configured
	if config.Global.BrevoAPIKey != "" {
		brevoConfig := &brevo.Config{
			APIKey:  config.Global.BrevoAPIKey,
			BaseURL: config.Global.BrevoBaseURL,
			Timeout: config.Global.BrevoTimeout,
		}

		brevoProvider := brevo.NewProvider(brevoConfig)
		if err := registry.Register(brevoProvider); err != nil {
			return nil, err
		}

		logger.Infof("registered Brevo provider")
	}

	// Set default provider if specified
	if config.Global.DefaultProvider != "" {
		if err := registry.SetDefault(config.Global.DefaultProvider); err != nil {
			return nil, err
		}
	}

	return NewServer(config.Global.SMTPPort, config.Global.MaxSize, authUsers, config.Global.AuthEnabled, registry), nil
}

// NewServer creates a new SMTP server
func NewServer(port string, maxMessageSize int64, authUsers map[string]string, authEnabled bool, registry *provider.Registry) *Server {
	var authHandler *AuthHandler
	if authEnabled && len(authUsers) > 0 {
		authHandler = NewAuthHandler(authUsers)
	}

	backend := NewBackend(maxMessageSize, authHandler, authEnabled, registry)

	s := smtp.NewServer(backend)
	s.Addr = ":" + port
	s.Domain = "localhost"
	s.MaxMessageBytes = maxMessageSize
	s.AllowInsecureAuth = true

	return &Server{
		server: s,
		addr:   s.Addr,
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

// Run starts the SMTP server, blocks until termination signal, then executes Shutdown
func (s *Server) Run() {
	if err := s.Start(); err != nil {
		logger.Fatal(err)
	}

	if config.Global.AuthEnabled {
		logger.Infof("SMTP server started on :%s with authentication enabled", config.Global.SMTPPort)
	} else {
		logger.Infof("SMTP server started on :%s with authentication disabled", config.Global.SMTPPort)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-c:
		logger.Infof("received signal: %v", sig)
	case <-ctx.Done():
	}

	logger.Info("shutting down server")
	if err := s.Shutdown(context.Background()); err != nil {
		logger.Errorf("error during shutdown: %v", err)
	}
}
