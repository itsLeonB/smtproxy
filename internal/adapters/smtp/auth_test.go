package smtp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAuthHandler(t *testing.T) {
	users := map[string]string{
		"user1": "pass1",
		"user2": "pass2",
	}
	
	handler := NewAuthHandler(users)
	
	assert.NotNil(t, handler)
	assert.Equal(t, users, handler.users)
}

func TestAuthHandler_AuthPlain_Success(t *testing.T) {
	users := map[string]string{
		"testuser": "testpass",
	}
	handler := NewAuthHandler(users)
	
	err := handler.AuthPlain(nil, "testuser", "testpass")
	assert.NoError(t, err)
}

func TestAuthHandler_AuthPlain_InvalidCredentials(t *testing.T) {
	users := map[string]string{
		"testuser": "testpass",
	}
	handler := NewAuthHandler(users)
	
	err := handler.AuthPlain(nil, "testuser", "wrongpass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthHandler_AuthLogin(t *testing.T) {
	users := map[string]string{
		"testuser": "testpass",
	}
	handler := NewAuthHandler(users)
	
	err := handler.AuthLogin(nil, "testuser", "testpass")
	assert.NoError(t, err)
}

func TestParseAuthPlain_Success(t *testing.T) {
	// Base64 encoded "\x00testuser\x00testpass"
	encoded := "AHRlc3R1c2VyAHRlc3RwYXNz"
	
	username, password, err := ParseAuthPlain(encoded)
	
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "testpass", password)
}

func TestParseAuthPlain_InvalidBase64(t *testing.T) {
	_, _, err := ParseAuthPlain("invalid-base64!")
	assert.Error(t, err)
}

func TestParseAuthPlain_InvalidFormat(t *testing.T) {
	// Base64 encoded "invalid"
	encoded := "aW52YWxpZA=="
	
	_, _, err := ParseAuthPlain(encoded)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid AUTH PLAIN format")
}
