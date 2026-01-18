package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/itsLeonB/smtproxy/internal/adapters/smtp"
	"github.com/itsLeonB/smtproxy/internal/core/config"
	"github.com/itsLeonB/smtproxy/internal/core/logger"
)

func main() {
	logger.Init("smtproxy")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	server := smtp.NewServer(config.Global.SMTPAddr, config.Global.MaxSize)

	if err := server.Start(); err != nil {
		logger.Fatal(err)
	}

	logger.Infof("SMTP server started on %s", config.Global.SMTPAddr)

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
