package smtp

import (
	"strings"
	"testing"

	"github.com/itsLeonB/smtproxy/internal/domain/service/parser"
	"github.com/stretchr/testify/assert"
)

func TestSession_AuthPlain_Success(t *testing.T) {
	users := map[string]string{"testuser": "testpass"}
	authHandler := NewAuthHandler(users)
	session := &Session{
		authHandler: authHandler,
		authEnabled: true,
	}
	
	err := session.AuthPlain("testuser", "testpass")
	assert.NoError(t, err)
	assert.NotNil(t, session.identity)
	assert.Equal(t, "testuser", session.identity.Username)
}

func TestSession_AuthPlain_Disabled(t *testing.T) {
	session := &Session{authEnabled: false}
	
	err := session.AuthPlain("testuser", "testpass")
	assert.NoError(t, err)
	assert.NotNil(t, session.identity)
	assert.Equal(t, "anonymous", session.identity.Username)
	assert.True(t, session.identity.IsAuthenticated())
}

func TestSession_AuthLogin_Disabled(t *testing.T) {
	session := &Session{authEnabled: false}
	
	err := session.AuthLogin("testuser", "testpass")
	assert.NoError(t, err)
	assert.NotNil(t, session.identity)
	assert.Equal(t, "anonymous", session.identity.Username)
	assert.True(t, session.identity.IsAuthenticated())
}

func TestSession_Mail_WithAuth(t *testing.T) {
	session := &Session{
		authEnabled: true,
		identity:    NewClientIdentity("testuser"),
	}
	
	err := session.Mail("test@example.com", nil)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", session.from)
}

func TestSession_Mail_NoAuth(t *testing.T) {
	session := &Session{authEnabled: true}
	
	err := session.Mail("test@example.com", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")
}

func TestSession_Mail_AuthDisabled(t *testing.T) {
	session := &Session{authEnabled: false}
	
	err := session.Mail("test@example.com", nil)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", session.from)
}

func TestSession_Rcpt_WithAuth(t *testing.T) {
	session := &Session{
		authEnabled: true,
		identity:    NewClientIdentity("testuser"),
	}
	
	err := session.Rcpt("user@example.com", nil)
	assert.NoError(t, err)
	assert.Contains(t, session.to, "user@example.com")
}

func TestSession_Data_WithAuth(t *testing.T) {
	session := &Session{
		from:           "sender@example.com",
		to:             []string{"recipient@example.com"},
		maxMessageSize: 1024,
		authEnabled:    true,
		identity:       NewClientIdentity("testuser"),
		parser:         parser.New(1024),
	}
	
	message := strings.NewReader("Subject: Test\n\nHello World")
	err := session.Data(message)
	assert.NoError(t, err)
}

func TestSession_Data_NoAuth(t *testing.T) {
	session := &Session{
		from:           "sender@example.com",
		to:             []string{"recipient@example.com"},
		maxMessageSize: 1024,
		authEnabled:    true,
		parser:         parser.New(1024),
	}
	
	message := strings.NewReader("Subject: Test\n\nHello World")
	err := session.Data(message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")
}

func TestSession_GetIdentity(t *testing.T) {
	identity := NewClientIdentity("testuser")
	session := &Session{identity: identity}
	
	result := session.GetIdentity()
	assert.Equal(t, identity, result)
}

// Keep existing tests for backward compatibility
func TestSession_Mail(t *testing.T) {
	session := &Session{authEnabled: false}
	err := session.Mail("test@example.com", nil)
	
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", session.from)
}

func TestSession_Rcpt(t *testing.T) {
	session := &Session{authEnabled: false}
	
	err := session.Rcpt("user1@example.com", nil)
	assert.NoError(t, err)
	
	err = session.Rcpt("user2@example.com", nil)
	assert.NoError(t, err)
	
	assert.Len(t, session.to, 2)
	assert.Contains(t, session.to, "user1@example.com")
	assert.Contains(t, session.to, "user2@example.com")
}

func TestSession_Data_WithEmailParsing(t *testing.T) {
	session := &Session{
		from:           "sender@example.com",
		to:             []string{"recipient@example.com"},
		maxMessageSize: 1024,
		authEnabled:    false,
		parser:         parser.New(1024),
	}
	
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: Test Email

Hello World!`
	
	err := session.Data(strings.NewReader(rawEmail))
	assert.NoError(t, err)
}

func TestSession_GetParsedEmail(t *testing.T) {
	session := &Session{
		parser: parser.New(1024),
	}
	
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: Test Email

Hello World!`
	
	email, err := session.GetParsedEmail(strings.NewReader(rawEmail))
	assert.NoError(t, err)
	assert.NotNil(t, email)
	assert.Equal(t, "sender@example.com", email.Headers.From.Address)
	assert.Equal(t, "Test Email", email.Headers.Subject)
	assert.Equal(t, "Hello World!", email.TextBody)
}

func TestSession_Data_NoSender(t *testing.T) {
	session := &Session{
		to:             []string{"recipient@example.com"},
		maxMessageSize: 1024,
		authEnabled:    false,
		parser:         parser.New(1024),
	}
	
	message := strings.NewReader("Subject: Test\n\nHello World")
	err := session.Data(message)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no sender specified")
}

func TestSession_Data_NoRecipients(t *testing.T) {
	session := &Session{
		from:           "sender@example.com",
		maxMessageSize: 1024,
		authEnabled:    false,
		parser:         parser.New(1024),
	}
	
	message := strings.NewReader("Subject: Test\n\nHello World")
	err := session.Data(message)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients specified")
}

func TestSession_Reset(t *testing.T) {
	session := &Session{
		from: "test@example.com",
		to:   []string{"user@example.com"},
	}
	
	session.Reset()
	
	assert.Empty(t, session.from)
	assert.Nil(t, session.to)
}

func TestSession_Data_ResetsStateAfterSuccess(t *testing.T) {
	session := &Session{
		from:           "sender@example.com",
		to:             []string{"recipient@example.com"},
		maxMessageSize: 1024,
		authEnabled:    false,
		parser:         parser.New(1024),
		dispatcher:     nil, // No dispatcher for this test
	}

	// Verify initial state
	assert.Equal(t, "sender@example.com", session.from)
	assert.Equal(t, []string{"recipient@example.com"}, session.to)

	// Call Data method with valid email
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: Test Email

Hello World!`
	
	err := session.Data(strings.NewReader(rawEmail))

	// Verify success and state reset
	assert.NoError(t, err)
	assert.Empty(t, session.from)
	assert.Nil(t, session.to)
}

func TestSession_Logout(t *testing.T) {
	session := &Session{}
	err := session.Logout()
	
	assert.NoError(t, err)
}
