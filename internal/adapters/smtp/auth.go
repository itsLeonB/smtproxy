package smtp

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/emersion/go-smtp"
)

// AuthHandler handles SMTP authentication
type AuthHandler struct {
	users map[string]string
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(users map[string]string) *AuthHandler {
	return &AuthHandler{
		users: users,
	}
}

// AuthPlain handles AUTH PLAIN mechanism
func (a *AuthHandler) AuthPlain(conn *smtp.Conn, username, password string) error {
	if expectedPassword, exists := a.users[username]; exists && expectedPassword == password {
		return nil
	}
	return errors.New("invalid credentials")
}

// AuthLogin handles AUTH LOGIN mechanism
func (a *AuthHandler) AuthLogin(conn *smtp.Conn, username, password string) error {
	return a.AuthPlain(conn, username, password)
}

// ParseAuthPlain parses AUTH PLAIN credentials from base64 encoded string
func ParseAuthPlain(encoded string) (username, password string, err error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}
	
	parts := strings.Split(string(decoded), "\x00")
	if len(parts) != 3 {
		return "", "", errors.New("invalid AUTH PLAIN format")
	}
	
	return parts[1], parts[2], nil
}
