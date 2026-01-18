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
}

// Mail handles MAIL FROM command
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	return nil
}

// Rcpt handles RCPT TO command
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

// Data handles DATA command
func (s *Session) Data(r io.Reader) error {
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
