package smtp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSession_Mail(t *testing.T) {
	session := &Session{}
	err := session.Mail("test@example.com", nil)
	
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", session.from)
}

func TestSession_Rcpt(t *testing.T) {
	session := &Session{}
	
	err := session.Rcpt("user1@example.com", nil)
	assert.NoError(t, err)
	
	err = session.Rcpt("user2@example.com", nil)
	assert.NoError(t, err)
	
	assert.Len(t, session.to, 2)
	assert.Contains(t, session.to, "user1@example.com")
	assert.Contains(t, session.to, "user2@example.com")
}

func TestSession_Data_Success(t *testing.T) {
	session := &Session{
		from:           "sender@example.com",
		to:             []string{"recipient@example.com"},
		maxMessageSize: 1024,
	}
	
	message := strings.NewReader("Subject: Test\n\nHello World")
	err := session.Data(message)
	
	assert.NoError(t, err)
}

func TestSession_Data_NoSender(t *testing.T) {
	session := &Session{
		to:             []string{"recipient@example.com"},
		maxMessageSize: 1024,
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

func TestSession_Logout(t *testing.T) {
	session := &Session{}
	err := session.Logout()
	
	assert.NoError(t, err)
}
