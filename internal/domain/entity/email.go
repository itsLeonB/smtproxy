package entity

import (
	"io"
	"time"
)

// Email represents a normalized email message
type Email struct {
	Headers     Headers
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
	RawSize     int64
}

// Headers contains normalized email headers
type Headers struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Date        time.Time
	MessageID   string
	ContentType string
	Custom      map[string][]string
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Size        int64
	Content     io.Reader
}

// Address represents an email address with optional display name
type Address struct {
	Email string
	Name  string
}
