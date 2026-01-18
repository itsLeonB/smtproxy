package main

import (
	"github.com/itsLeonB/smtproxy/internal/adapters/smtp"
	"github.com/itsLeonB/smtproxy/internal/core/config"
	"github.com/itsLeonB/smtproxy/internal/core/logger"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init("smtproxy")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	srv, err := smtp.Setup()
	if err != nil {
		logger.Fatal(err)
	}

	srv.Run()
}
