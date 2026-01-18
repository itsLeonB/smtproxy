package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itsLeonB/smtproxy/internal/adapters/providers/brevo"
	"github.com/itsLeonB/smtproxy/internal/adapters/smtp"
	"github.com/itsLeonB/smtproxy/internal/core/config"
	"github.com/itsLeonB/smtproxy/internal/core/logger"
	"github.com/itsLeonB/smtproxy/internal/domain/service/provider"
)

func main() {
	logger.Init("smtproxy")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	authUsers := config.Global.AuthUsers

	// Initialize provider registry
	registry := provider.NewRegistry()

	// Register Brevo provider if configured
	if config.Global.BrevoAPIKey != "" {
		timeout, err := time.ParseDuration(config.Global.BrevoTimeout)
		if err != nil {
			logger.Fatal(err)
		}

		brevoConfig := &brevo.Config{
			APIKey:  config.Global.BrevoAPIKey,
			BaseURL: config.Global.BrevoBaseURL,
			Timeout: timeout,
		}

		brevoProvider := brevo.NewProvider(brevoConfig)
		if err := registry.Register(brevoProvider); err != nil {
			logger.Fatal(err)
		}

		logger.Infof("registered Brevo provider")
	}

	// Set default provider if specified
	if config.Global.DefaultProvider != "" {
		if err := registry.SetDefault(config.Global.DefaultProvider); err != nil {
			logger.Warnf("failed to set default provider %s: %v", config.Global.DefaultProvider, err)
		}
	}

	server := smtp.NewServer(config.Global.SMTPAddr, config.Global.MaxSize, authUsers, config.Global.AuthEnabled, registry, logger.Global)

	if err := server.Start(); err != nil {
		logger.Fatal(err)
	}

	if config.Global.AuthEnabled {
		logger.Infof("SMTP server started on %s with authentication enabled (%d users)", config.Global.SMTPAddr, len(authUsers))
	} else {
		logger.Infof("SMTP server started on %s with authentication disabled", config.Global.SMTPAddr)
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
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Errorf("error during shutdown: %v", err)
	}
}
