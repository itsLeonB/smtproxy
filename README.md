# smtproxy

A production-grade SMTP proxy that converts incoming SMTP messages to transactional email API calls. Built with Go for high performance and reliability.

## Features

- **SMTP Server** - Full SMTP protocol support with authentication
- **Email Parsing** - RFC-compliant MIME parsing with UTF-8 support
- **Provider Abstraction** - Pluggable transactional email providers
- **Authentication** - SMTP AUTH PLAIN/LOGIN with configurable users
- **Graceful Shutdown** - Signal handling with proper resource cleanup
- **Structured Logging** - Comprehensive logging with configurable levels

## Architecture

```
SMTP Client → SMTP Server → Email Parser → Dispatcher → Provider API
```

### Components

- **SMTP Adapter** - Handles SMTP protocol and authentication
- **Email Parser** - Converts raw SMTP DATA to normalized Email entities
- **Dispatcher** - Routes emails to providers with error translation
- **Provider Registry** - Manages multiple provider implementations
- **Providers** - HTTP clients for transactional email APIs

## Quick Start

### Installation

```bash
git clone https://github.com/itsLeonB/smtproxy
cd smtproxy
go build -o bin/smtproxy ./cmd/smtp
```

### Basic Usage

```bash
# Start with default settings
./bin/smtproxy

# Configure via environment variables
export SMTP_ADDR=":2525"
export AUTH_USERS="user1:pass1,user2:pass2"
export BREVO_API_KEY="your-brevo-api-key"
./bin/smtproxy

# Or use .env file
cp .env.example .env
# Edit .env with your configuration
./bin/smtproxy
```

### Send Test Email

```bash
# Using telnet
telnet localhost 2525
EHLO localhost
AUTH PLAIN AHVzZXIxAHBhc3Mx  # base64: \0user1\0pass1
MAIL FROM:<sender@example.com>
RCPT TO:<recipient@example.com>
DATA
Subject: Test Email

Hello World!
.
QUIT
```

## Configuration

All configuration is done via environment variables or `.env` file:

### Using .env File

```bash
# Copy example configuration
cp .env.example .env

# Edit configuration
nano .env

# Start application (automatically loads .env)
./bin/smtproxy
```

### Core Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `SMTP_ADDR` | `:2525` | SMTP server listen address |
| `MAX_MESSAGE_SIZE` | `10485760` | Maximum message size in bytes (10MB) |

### Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_ENABLED` | `true` | Enable SMTP authentication |
| `AUTH_USERS` | `user1:pass1,user2:pass2` | Comma-separated user:password pairs |

### Provider Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DEFAULT_PROVIDER` | `brevo` | Default provider to use |
| `ENABLED_PROVIDERS` | `brevo` | Comma-separated list of enabled providers |

### Brevo Provider

| Variable | Default | Description |
|----------|---------|-------------|
| `BREVO_API_KEY` | - | Brevo API key (required) |
| `BREVO_BASE_URL` | `https://api.brevo.com/v3` | Brevo API base URL |
| `BREVO_TIMEOUT` | `30s` | HTTP request timeout |

## Providers

### Brevo (Sendinblue)

The Brevo provider supports:
- HTML and plain text emails
- Multiple recipients (To, CC, BCC)
- Proper error mapping
- Rate limit handling

**Setup:**
1. Get API key from [Brevo Console](https://app.brevo.com/)
2. Set `BREVO_API_KEY` environment variable
3. Start smtproxy

**Error Mapping:**
- `400` → Invalid email address
- `401` → Authentication failed
- `402` → Insufficient credits
- `429` → Rate limit exceeded
- `5xx` → Service unavailable

## Development

### Project Structure

```
smtproxy/
├── cmd/smtp/                    # Application entry point
├── internal/
│   ├── adapters/
│   │   ├── providers/brevo/     # Brevo provider implementation
│   │   └── smtp/                # SMTP protocol adapter
│   ├── core/
│   │   ├── config/              # Configuration management
│   │   └── logger/              # Logging utilities
│   └── domain/
│       ├── entity/              # Domain entities (Email, etc.)
│       └── service/
│           ├── dispatcher/      # Email dispatch logic
│           ├── parser/          # MIME email parsing
│           └── provider/        # Provider abstraction
└── bin/                         # Compiled binaries
```

### Building

```bash
# Build for current platform
go build -o bin/smtproxy ./cmd/smtp

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/smtproxy-linux ./cmd/smtp

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o bin/smtproxy.exe ./cmd/smtp
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/adapters/providers/brevo -v
```

### Adding New Providers

1. Create provider package in `internal/adapters/providers/`
2. Implement the `Provider` interface:
   ```go
   type Provider interface {
       Name() string
       Send(ctx context.Context, email *entity.Email) error
       IsHealthy(ctx context.Context) error
   }
   ```
3. Add configuration to `internal/core/config/`
4. Register provider in `cmd/smtp/main.go`
5. Add comprehensive tests

## Deployment

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o smtproxy ./cmd/smtp

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/smtproxy .
CMD ["./smtproxy"]
```

### Systemd Service

```ini
[Unit]
Description=SMTP Proxy
After=network.target

[Service]
Type=simple
User=smtproxy
WorkingDirectory=/opt/smtproxy
ExecStart=/opt/smtproxy/bin/smtproxy
Restart=always
RestartSec=5
Environment=SMTP_ADDR=:2525
Environment=BREVO_API_KEY=your-api-key

[Install]
WantedBy=multi-user.target
```

### Environment Variables File

```bash
# .env
LOG_LEVEL=info
SMTP_ADDR=:2525
MAX_MESSAGE_SIZE=10485760
AUTH_ENABLED=true
AUTH_USERS=user1:pass1,user2:pass2
DEFAULT_PROVIDER=brevo
BREVO_API_KEY=your-brevo-api-key
BREVO_TIMEOUT=30s
```

## Monitoring

### Health Check

The application provides basic health monitoring through logs:

```bash
# Check if server started successfully
tail -f /var/log/smtproxy.log | grep "SMTP server started"

# Monitor email dispatch
tail -f /var/log/smtproxy.log | grep "email dispatched"

# Monitor errors
tail -f /var/log/smtproxy.log | grep "ERROR"
```

### Metrics

Key metrics to monitor:
- SMTP connections accepted/rejected
- Authentication success/failure rates
- Email dispatch success/failure rates
- Provider API response times
- Error rates by provider

## Security

### Authentication

- SMTP AUTH PLAIN and LOGIN supported
- Configurable user credentials
- Authentication required by default
- Failed authentication attempts logged

### Network Security

- SMTP server binds to configurable address
- HTTPS-only provider API calls
- No sensitive data in logs
- Graceful handling of malformed requests

### Best Practices

1. Use strong passwords for SMTP users
2. Rotate API keys regularly
3. Monitor authentication failures
4. Use TLS for SMTP connections in production
5. Limit message size to prevent abuse

## Troubleshooting

### Common Issues

**SMTP Connection Refused**
```bash
# Check if server is running
netstat -tlnp | grep :2525

# Check logs for startup errors
tail -f /var/log/smtproxy.log
```

**Authentication Failed**
```bash
# Verify user credentials
echo -n '\0user1\0pass1' | base64  # Should match AUTH_USERS

# Check authentication logs
grep "authentication" /var/log/smtproxy.log
```

**Provider API Errors**
```bash
# Check API key configuration
echo $BREVO_API_KEY

# Test provider health
curl -H "api-key: $BREVO_API_KEY" https://api.brevo.com/v3/account
```

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
LOG_LEVEL=debug ./bin/smtproxy
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Code Style

- Follow Go conventions
- Add comprehensive tests
- Document public APIs
- Use structured logging
- Handle errors gracefully

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- GitHub Issues: [Report bugs or request features](https://github.com/itsLeonB/smtproxy/issues)
- Documentation: This README and inline code comments
- Examples: See `examples/` directory for usage examples