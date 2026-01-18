package entity

import (
	"io"
	"net/mail"
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
	From        *mail.Address
	To          []*mail.Address
	CC          []*mail.Address
	BCC         []*mail.Address
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
