package parser

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParseSimpleEmail(t *testing.T) {
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: Test Subject
Date: Mon, 02 Jan 2006 15:04:05 -0700
Content-Type: text/plain

Hello World!`

	parser := New(1024 * 1024)
	email, err := parser.Parse(strings.NewReader(rawEmail))

	assert.NoError(t, err)
	assert.NotNil(t, email)
	assert.Equal(t, "sender@example.com", email.Headers.From.Address)
	assert.Equal(t, "", email.Headers.From.Name)
	assert.Len(t, email.Headers.To, 1)
	assert.Equal(t, "recipient@example.com", email.Headers.To[0].Address)
	assert.Equal(t, "", email.Headers.To[0].Name)
	assert.Equal(t, "Test Subject", email.Headers.Subject)
	assert.Equal(t, "Hello World!", email.TextBody)
	assert.Empty(t, email.HTMLBody)
	assert.Empty(t, email.Attachments)
}

func TestParser_ParseMultipartEmail(t *testing.T) {
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: Multipart Test
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain

Plain text content
--boundary123
Content-Type: text/html

<html><body>HTML content</body></html>
--boundary123--`

	parser := New(1024 * 1024)
	email, err := parser.Parse(strings.NewReader(rawEmail))

	assert.NoError(t, err)
	assert.NotNil(t, email)
	assert.Equal(t, "Plain text content", strings.TrimSpace(email.TextBody))
	assert.Equal(t, "<html><body>HTML content</body></html>", strings.TrimSpace(email.HTMLBody))
}

func TestParser_ParseHeaders(t *testing.T) {
	rawEmail := `From: "John Doe" <john@example.com>
To: user1@example.com, user2@example.com
CC: cc@example.com
Subject: =?UTF-8?B?VGVzdCBTdWJqZWN0?=
Date: Mon, 02 Jan 2006 15:04:05 -0700
Message-ID: <123@example.com>
X-Custom-Header: custom value

Test body`

	parser := New(1024 * 1024)
	email, err := parser.Parse(strings.NewReader(rawEmail))

	assert.NoError(t, err)
	assert.Equal(t, "john@example.com", email.Headers.From.Address)
	assert.Equal(t, "John Doe", email.Headers.From.Name)
	assert.Len(t, email.Headers.To, 2)
	assert.Equal(t, "user1@example.com", email.Headers.To[0].Address)
	assert.Equal(t, "user2@example.com", email.Headers.To[1].Address)
	assert.Len(t, email.Headers.CC, 1)
	assert.Equal(t, "cc@example.com", email.Headers.CC[0].Address)
	assert.Equal(t, "Test Subject", email.Headers.Subject)
	assert.Equal(t, "<123@example.com>", email.Headers.MessageID)
	assert.Contains(t, email.Headers.Custom, "X-Custom-Header")
}

func TestParser_ParseDate(t *testing.T) {
	rawEmail := `From: sender@example.com
To: recipient@example.com
Date: Mon, 02 Jan 2006 15:04:05 -0700

Test body`

	parser := New(1024 * 1024)
	email, err := parser.Parse(strings.NewReader(rawEmail))

	assert.NoError(t, err)
	expected := time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", -7*3600))
	assert.Equal(t, expected, email.Headers.Date)
}

func TestParser_DecodeHeader(t *testing.T) {
	parser := New(1024)
	
	// Test RFC 2047 encoded header
	encoded := "=?UTF-8?B?VGVzdCBTdWJqZWN0?="
	decoded := parser.decodeHeader(encoded)
	assert.Equal(t, "Test Subject", decoded)
	
	// Test plain header
	plain := "Plain Subject"
	decoded = parser.decodeHeader(plain)
	assert.Equal(t, "Plain Subject", decoded)
}

func TestParser_ParseAddress(t *testing.T) {
	parser := New(1024)
	
	tests := []struct {
		input         string
		expectedEmail string
		expectedName  string
		expectNil     bool
	}{
		{"ellionblessan@gmail.com", "ellionblessan@gmail.com", "", false},
		{"FOSS Sure <ellionblessan@gmail.com>", "ellionblessan@gmail.com", "FOSS Sure", false},
		{"John Doe <john@example.com>", "john@example.com", "John Doe", false},
		{"<test@example.com>", "test@example.com", "", false},
		{"", "", "", true},  // Empty string returns nil
		{"invalid-email", "invalid-email", "", false},  // Fallback to original
	}
	
	for _, tt := range tests {
		result := parser.parseAddress(tt.input)
		if tt.expectNil {
			assert.Nil(t, result, "Expected nil for input: %s", tt.input)
		} else {
			assert.NotNil(t, result, "Expected non-nil for input: %s", tt.input)
			assert.Equal(t, tt.expectedEmail, result.Address, "Failed email for input: %s", tt.input)
			assert.Equal(t, tt.expectedName, result.Name, "Failed name for input: %s", tt.input)
		}
	}
}

func TestParser_ParseAddressList(t *testing.T) {
	parser := New(1024)
	
	addresses := "user1@example.com, John Doe <user2@example.com>"
	result := parser.parseAddressList(addresses)
	
	assert.Len(t, result, 2)
	assert.Equal(t, "user1@example.com", result[0].Address)
	assert.Equal(t, "", result[0].Name)
	assert.Equal(t, "user2@example.com", result[1].Address)
	assert.Equal(t, "John Doe", result[1].Name)
}

func TestParser_IsStandardHeader(t *testing.T) {
	parser := New(1024)
	
	assert.True(t, parser.isStandardHeader("From"))
	assert.True(t, parser.isStandardHeader("subject"))
	assert.False(t, parser.isStandardHeader("X-Custom"))
}

func TestParser_DecodeContent_Base64(t *testing.T) {
	parser := New(1024)
	
	// "Hello World!" in base64
	encoded := "SGVsbG8gV29ybGQh"
	content, err := parser.decodeContent(strings.NewReader(encoded), "base64")
	
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", string(content))
}

func TestParser_DecodeContent_QuotedPrintable(t *testing.T) {
	parser := New(1024)
	
	// "Hello World!" in quoted-printable
	encoded := "Hello=20World!"
	content, err := parser.decodeContent(strings.NewReader(encoded), "quoted-printable")
	
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", string(content))
}

func TestParser_DecodeContent_PlainText(t *testing.T) {
	parser := New(1024)
	
	tests := []string{"7bit", "8bit", "binary", ""}
	
	for _, encoding := range tests {
		content, err := parser.decodeContent(strings.NewReader("Hello World!"), encoding)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World!", string(content))
	}
}

func TestParser_ParseEmailWithBase64Content(t *testing.T) {
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: Base64 Test
Content-Type: text/plain
Content-Transfer-Encoding: base64

SGVsbG8gV29ybGQh`

	parser := New(1024 * 1024)
	email, err := parser.Parse(strings.NewReader(rawEmail))

	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", strings.TrimSpace(email.TextBody))
}

func TestParser_ParseMultipartWithEncodedParts(t *testing.T) {
	rawEmail := `From: sender@example.com
To: recipient@example.com
Subject: Encoded Multipart Test
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain
Content-Transfer-Encoding: base64

SGVsbG8gV29ybGQh
--boundary123
Content-Type: text/html
Content-Transfer-Encoding: quoted-printable

<html><body>Hello=20World!</body></html>
--boundary123--`

	parser := New(1024 * 1024)
	email, err := parser.Parse(strings.NewReader(rawEmail))

	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", strings.TrimSpace(email.TextBody))
	assert.Equal(t, "<html><body>Hello World!</body></html>", strings.TrimSpace(email.HTMLBody))
}
