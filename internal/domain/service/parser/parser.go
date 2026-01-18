package parser

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/itsLeonB/smtproxy/internal/core/logger"
	"github.com/itsLeonB/smtproxy/internal/domain/entity"
)

// Parser handles MIME email parsing
type Parser struct {
	maxSize int64
}

// New creates a new email parser
func New(maxSize int64) *Parser {
	return &Parser{maxSize: maxSize}
}

// Parse converts raw email data into normalized Email model
func (p *Parser) Parse(r io.Reader) (*entity.Email, error) {
	// Limit reader to prevent memory exhaustion
	lr := io.LimitReader(r, p.maxSize)

	// Parse message
	msg, err := mail.ReadMessage(lr)
	if err != nil {
		return nil, err
	}

	// Extract headers
	headers := p.parseHeaders(msg.Header)

	// Parse body based on content type
	contentType := msg.Header.Get("Content-Type")
	mediaType, params, _ := mime.ParseMediaType(contentType)

	parsedEmail := &entity.Email{
		Headers: headers,
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		err = p.parseMultipart(msg.Body, params["boundary"], parsedEmail)
	} else {
		err = p.parseSinglePart(msg.Body, mediaType, parsedEmail, msg.Header)
	}

	return parsedEmail, err
}

// parseHeaders extracts and normalizes email headers
func (p *Parser) parseHeaders(h mail.Header) entity.Headers {
	headers := entity.Headers{
		Custom: make(map[string][]string),
	}

	// Standard headers
	headers.From = p.parseAddress(p.decodeHeader(h.Get("From")))
	headers.Subject = p.decodeHeader(h.Get("Subject"))
	headers.MessageID = h.Get("Message-ID")
	headers.ContentType = h.Get("Content-Type")

	// Parse date
	if dateStr := h.Get("Date"); dateStr != "" {
		if date, err := mail.ParseDate(dateStr); err == nil {
			headers.Date = date
		}
	}

	// Parse address lists
	headers.To = p.parseAddressList(h.Get("To"))
	headers.CC = p.parseAddressList(h.Get("CC"))
	headers.BCC = p.parseAddressList(h.Get("BCC"))

	// Store custom headers
	for key, values := range h {
		if !p.isStandardHeader(key) {
			headers.Custom[key] = values
		}
	}

	return headers
}

// parseMultipart handles multipart MIME messages
func (p *Parser) parseMultipart(body io.Reader, boundary string, msg *entity.Email) error {
	mr := multipart.NewReader(body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := p.processPart(part, msg); err != nil {
			if e := part.Close(); e != nil {
				logger.Error(e)
			}
			return err
		}
		if e := part.Close(); e != nil {
			logger.Error(e)
		}
	}

	return nil
}

// processPart handles individual MIME parts
func (p *Parser) processPart(part *multipart.Part, msg *entity.Email) error {
	contentType := part.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(contentType)

	disposition := part.Header.Get("Content-Disposition")
	dispType, dispParams, _ := mime.ParseMediaType(disposition)

	encoding := part.Header.Get("Content-Transfer-Encoding")

	// Handle attachments
	if dispType == "attachment" || dispParams["filename"] != "" {
		return p.processAttachment(part, msg, dispParams, encoding)
	}

	// Handle body content
	switch mediaType {
	case "text/plain":
		content, err := p.decodeContent(part, encoding)
		if err != nil {
			return err
		}
		msg.TextBody = string(content)
	case "text/html":
		content, err := p.decodeContent(part, encoding)
		if err != nil {
			return err
		}
		msg.HTMLBody = string(content)
	case "multipart/alternative", "multipart/mixed":
		// Nested multipart - would need recursive handling
		// For minimal implementation, skip
	}

	return nil
}

// processAttachment handles file attachments
func (p *Parser) processAttachment(part *multipart.Part, msg *entity.Email, params map[string]string, encoding string) error {
	filename := params["filename"]
	if filename == "" {
		filename = "attachment"
	}

	// Decode filename if encoded
	filename = p.decodeHeader(filename)

	// Read content into buffer for size calculation
	content, err := p.decodeContent(part, encoding)
	if err != nil {
		return err
	}

	attachment := entity.Attachment{
		Filename:    filename,
		ContentType: part.Header.Get("Content-Type"),
		Size:        int64(len(content)),
		Content:     bytes.NewReader(content),
	}

	msg.Attachments = append(msg.Attachments, attachment)
	return nil
}

// parseSinglePart handles non-multipart messages
func (p *Parser) parseSinglePart(body io.Reader, mediaType string, msg *entity.Email, headers mail.Header) error {
	encoding := headers.Get("Content-Transfer-Encoding")
	content, err := p.decodeContent(body, encoding)
	if err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(mediaType, "text/plain"):
		msg.TextBody = string(content)
	case strings.HasPrefix(mediaType, "text/html"):
		msg.HTMLBody = string(content)
	default:
		msg.TextBody = string(content)
	}

	return nil
}

// decodeContent handles content transfer encoding
func (p *Parser) decodeContent(r io.Reader, encoding string) ([]byte, error) {
	encoding = strings.ToLower(strings.TrimSpace(encoding))
	
	switch encoding {
	case "base64":
		decoder := base64.NewDecoder(base64.StdEncoding, r)
		return io.ReadAll(decoder)
	case "quoted-printable":
		decoder := quotedprintable.NewReader(r)
		return io.ReadAll(decoder)
	case "7bit", "8bit", "binary", "":
		return io.ReadAll(r)
	default:
		// Unknown encoding, fallback to raw read
		return io.ReadAll(r)
	}
}

// decodeHeader decodes RFC 2047 encoded headers
func (p *Parser) decodeHeader(header string) string {
	decoder := &mime.WordDecoder{}
	decoded, err := decoder.DecodeHeader(header)
	if err != nil {
		return header // Return original if decode fails
	}
	return decoded
}

// parseAddress parses a single email address
func (p *Parser) parseAddress(address string) *mail.Address {
	if address == "" {
		return nil
	}

	if addr, err := mail.ParseAddress(address); err == nil {
		return addr
	}

	// Fallback to plain email
	return &mail.Address{
		Address: strings.TrimSpace(address),
	}
}

// parseAddressList parses comma-separated email addresses
func (p *Parser) parseAddressList(addresses string) []*mail.Address {
	if addresses == "" {
		return nil
	}

	addrs, err := mail.ParseAddressList(addresses)
	if err != nil {
		// Fallback to simple split if parsing fails
		parts := strings.Split(addresses, ",")
		result := make([]*mail.Address, 0, len(parts))
		for _, part := range parts {
			if addr := p.parseAddress(strings.TrimSpace(part)); addr != nil && addr.Address != "" {
				result = append(result, addr)
			}
		}
		return result
	}

	return addrs
}

// isStandardHeader checks if header is a standard email header
func (p *Parser) isStandardHeader(key string) bool {
	standard := []string{
		"From", "To", "CC", "BCC", "Subject", "Date",
		"Message-ID", "Content-Type", "Content-Disposition",
	}

	key = strings.ToLower(key)
	for _, std := range standard {
		if strings.ToLower(std) == key {
			return true
		}
	}
	return false
}
