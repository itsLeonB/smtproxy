package smtp

// ClientIdentity represents an authenticated SMTP client
type ClientIdentity struct {
	Username      string
	Authenticated bool
}

// NewClientIdentity creates a new client identity
func NewClientIdentity(username string) *ClientIdentity {
	return &ClientIdentity{
		Username:      username,
		Authenticated: true,
	}
}

// IsAuthenticated returns whether the client is authenticated
func (c *ClientIdentity) IsAuthenticated() bool {
	return c.Authenticated
}
