package smtp

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/emersion/go-smtp"
	"github.com/itsLeonB/smtproxy/internal/domain/entity"
	"github.com/itsLeonB/smtproxy/internal/domain/service/dispatcher"
	"github.com/itsLeonB/smtproxy/internal/domain/service/parser"
)

// Session implements smtp.Session interface
type Session struct {
	from           string
	to             []string
	maxMessageSize int64
	authHandler    *AuthHandler
	authEnabled    bool
	identity       *ClientIdentity
	parser         *parser.Parser
	dispatcher     *dispatcher.Dispatcher
}

// AuthPlain handles AUTH PLAIN authentication
func (s *Session) AuthPlain(username, password string) error {
	if !s.authEnabled {
		// Allow anonymous authentication when auth is disabled
		s.identity = NewClientIdentity("anonymous")
		return nil
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
		// Allow anonymous authentication when auth is disabled
		s.identity = NewClientIdentity("anonymous")
		return nil
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

	// Validate message size
	limitedReader := &sizeLimitReader{
		reader:   r,
		maxSize:  s.maxMessageSize,
		bytesRead: 0,
	}

	// Parse email using the MIME parser
	parsedEmail, err := s.parser.Parse(limitedReader)
	if err != nil {
		return err
	}

	// Dispatch email via provider
	if s.dispatcher != nil {
		err = s.dispatcher.Dispatch(context.Background(), parsedEmail, "")
		if err != nil {
			return err
		}
	}

	// Reset session state after successful processing
	s.Reset()
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

// GetParsedEmail returns the last parsed email (for testing)
func (s *Session) GetParsedEmail(r io.Reader) (*entity.Email, error) {
	return s.parser.Parse(r)
}

// sizeLimitReader wraps an io.Reader and enforces a size limit
type sizeLimitReader struct {
	reader    io.Reader
	maxSize   int64
	bytesRead int64
}

func (r *sizeLimitReader) Read(p []byte) (n int, err error) {
	if r.bytesRead >= r.maxSize {
		return 0, fmt.Errorf("message size exceeds maximum allowed size of %d bytes", r.maxSize)
	}

	// Limit the read to not exceed maxSize
	remaining := r.maxSize - r.bytesRead
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err = r.reader.Read(p)
	r.bytesRead += int64(n)

	// Check if we've exceeded the limit after reading
	if r.bytesRead > r.maxSize {
		return n, fmt.Errorf("message size exceeds maximum allowed size of %d bytes", r.maxSize)
	}

	return n, err
}
