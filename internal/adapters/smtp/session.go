package smtp

import (
	"errors"
	"io"

	"github.com/emersion/go-smtp"
)

// Session implements smtp.Session interface
type Session struct {
	from           string
	to             []string
	maxMessageSize int64
	authHandler    *AuthHandler
	authEnabled    bool
	identity       *ClientIdentity
}

// AuthPlain handles AUTH PLAIN authentication
func (s *Session) AuthPlain(username, password string) error {
	if !s.authEnabled {
		return errors.New("authentication disabled")
	}
	
	if s.authHandler == nil {
		return errors.New("no auth handler configured")
	}
	
	if err := s.authHandler.AuthPlain(nil, username, password); err != nil {
		return err
	}
	
	s.identity = NewClientIdentity(username)
	return nil
}

// AuthLogin handles AUTH LOGIN authentication
func (s *Session) AuthLogin(username, password string) error {
	if !s.authEnabled {
		return errors.New("authentication disabled")
	}
	
	if s.authHandler == nil {
		return errors.New("no auth handler configured")
	}
	
	if err := s.authHandler.AuthLogin(nil, username, password); err != nil {
		return err
	}
	
	s.identity = NewClientIdentity(username)
	return nil
}

// Mail handles MAIL FROM command
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	if s.authEnabled && (s.identity == nil || !s.identity.IsAuthenticated()) {
		return errors.New("authentication required")
	}
	
	s.from = from
	return nil
}

// Rcpt handles RCPT TO command
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	if s.authEnabled && (s.identity == nil || !s.identity.IsAuthenticated()) {
		return errors.New("authentication required")
	}
	
	s.to = append(s.to, to)
	return nil
}

// Data handles DATA command
func (s *Session) Data(r io.Reader) error {
	if s.authEnabled && (s.identity == nil || !s.identity.IsAuthenticated()) {
		return errors.New("authentication required")
	}
	
	if s.from == "" {
		return errors.New("no sender specified")
	}
	if len(s.to) == 0 {
		return errors.New("no recipients specified")
	}

	// Read message with size limit
	lr := io.LimitReader(r, s.maxMessageSize)
	_, err := io.ReadAll(lr)
	if err != nil {
		return err
	}

	return nil
}

// Reset resets the session state
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

// Logout handles session cleanup
func (s *Session) Logout() error {
	return nil
}

// GetIdentity returns the client identity
func (s *Session) GetIdentity() *ClientIdentity {
	return s.identity
}
