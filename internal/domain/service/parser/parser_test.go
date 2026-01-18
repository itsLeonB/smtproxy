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
	assert.Equal(t, "sender@example.com", email.Headers.From)
	assert.Equal(t, []string{"recipient@example.com"}, email.Headers.To)
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
	assert.Equal(t, `"John Doe" <john@example.com>`, email.Headers.From)
	assert.Equal(t, []string{"user1@example.com", "user2@example.com"}, email.Headers.To)
	assert.Equal(t, []string{"cc@example.com"}, email.Headers.CC)
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

func TestParser_ParseAddressList(t *testing.T) {
	parser := New(1024)
	
	addresses := "user1@example.com, user2@example.com"
	result := parser.parseAddressList(addresses)
	
	assert.Equal(t, []string{"user1@example.com", "user2@example.com"}, result)
}

func TestParser_IsStandardHeader(t *testing.T) {
	parser := New(1024)
	
	assert.True(t, parser.isStandardHeader("From"))
	assert.True(t, parser.isStandardHeader("subject"))
	assert.False(t, parser.isStandardHeader("X-Custom"))
}
