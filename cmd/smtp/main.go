package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
	// TODO: Register actual providers based on config
	
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
