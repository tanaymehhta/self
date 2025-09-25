# Self Integrations

Microservices for connecting with external platforms and services.

## Overview

The integrations hub consists of specialized microservices that handle OAuth authentication, webhook processing, and data synchronization with various external platforms.

## Architecture

Each integration is a standalone microservice that:
- Handles OAuth 2.0 authentication flows
- Processes webhooks from external services
- Normalizes data into common formats
- Provides consistent APIs for the backend

## Supported Integrations

### Communication
- **Google Gmail** - Email thread analysis and correlation
- **Microsoft Outlook** - Calendar and email integration
- **Slack** - Channel and DM message correlation
- **Discord** - Server message and voice chat integration
- **Microsoft Teams** - Meeting and chat analysis
- **Zoom** - Meeting transcription and recording

### Productivity
- **Google Calendar** - Meeting and event correlation
- **Google Drive** - File activity and content analysis
- **Dropbox** - File synchronization and sharing events
- **OneDrive** - Document collaboration tracking
- **Notion** - Page and database updates
- **Obsidian** - Note-taking and knowledge graph

### Development
- **GitHub** - Repository activity and issue tracking
- **GitLab** - Merge request and pipeline correlation
- **Linear** - Issue and project management
- **Jira** - Ticket and project tracking
- **Figma** - Design collaboration events

### Social & Professional
- **LinkedIn** - Professional network updates
- **Twitter** - Social context and mentions
- **Calendar** - Meeting and event correlation

## Project Structure

```
integrations/
├── shared/               # Shared utilities and types
│   ├── oauth/           # OAuth 2.0 handlers
│   ├── webhook/         # Webhook processing
│   ├── types/           # Common data models
│   └── utils/           # Helper functions
├── google/              # Google services (Gmail, Calendar, Drive)
├── microsoft/           # Microsoft services (Outlook, Teams, OneDrive)
├── slack/               # Slack integration
├── github/              # GitHub integration
├── notion/              # Notion integration
└── docker-compose.yml   # Development environment
```

## Service Template

Each integration follows a common structure:

```
service-name/
├── cmd/
│   └── main.go          # Service entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── auth/            # OAuth implementation
│   ├── client/          # API client
│   ├── models/          # Data models
│   ├── service/         # Business logic
│   └── webhook/         # Webhook handlers
├── configs/
│   └── config.yaml      # Service configuration
├── Dockerfile
└── README.md
```

## Shared Components

### OAuth Handler

```go
package oauth

type Provider struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
    AuthURL      string
    TokenURL     string
}

func (p *Provider) GetAuthURL(state string) string {
    // Generate OAuth authorization URL
}

func (p *Provider) ExchangeCodeForToken(code string) (*Token, error) {
    // Exchange authorization code for access token
}

func (p *Provider) RefreshToken(refreshToken string) (*Token, error) {
    // Refresh access token using refresh token
}
```

### Webhook Processor

```go
package webhook

type Processor struct {
    Secret    string
    Handlers  map[string]Handler
}

type Event struct {
    Type      string                 `json:"type"`
    Source    string                 `json:"source"`
    Timestamp time.Time             `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
    UserID    string                 `json:"user_id"`
}

func (p *Processor) ProcessWebhook(payload []byte, signature string) error {
    // Verify webhook signature and process event
}
```

### Data Normalizer

```go
package normalizer

type Message struct {
    ID          string            `json:"id"`
    Source      string            `json:"source"`
    Type        string            `json:"type"`
    Content     string            `json:"content"`
    Author      Person            `json:"author"`
    Timestamp   time.Time         `json:"timestamp"`
    ThreadID    string            `json:"thread_id,omitempty"`
    Mentions    []Person          `json:"mentions,omitempty"`
    Attachments []Attachment      `json:"attachments,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type Person struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email,omitempty"`
}

type Attachment struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    URL      string `json:"url"`
    Type     string `json:"type"`
    Size     int64  `json:"size"`
}
```

## Integration Examples

### Google Gmail Integration

```go
package gmail

type Service struct {
    client   *gmail.Service
    oauth    *oauth.Provider
    webhook  *webhook.Processor
}

func (s *Service) ProcessEmail(email *gmail.Message) (*normalizer.Message, error) {
    // Convert Gmail message to normalized format
    return &normalizer.Message{
        ID:        email.Id,
        Source:    "gmail",
        Type:      "email",
        Content:   extractEmailContent(email),
        Author:    extractSender(email),
        Timestamp: parseTimestamp(email),
        ThreadID:  email.ThreadId,
    }, nil
}

func (s *Service) HandleWebhook(payload []byte) error {
    // Process Gmail push notification
    var notification GmailNotification
    if err := json.Unmarshal(payload, &notification); err != nil {
        return err
    }

    // Fetch and process new/modified messages
    return s.syncMessages(notification.HistoryId)
}
```

### Slack Integration

```go
package slack

type Service struct {
    client  *slack.Client
    oauth   *oauth.Provider
    webhook *webhook.Processor
}

func (s *Service) ProcessMessage(event slack.MessageEvent) (*normalizer.Message, error) {
    return &normalizer.Message{
        ID:        event.Timestamp,
        Source:    "slack",
        Type:      "message",
        Content:   event.Text,
        Author:    normalizer.Person{
            ID:   event.User,
            Name: s.getUserName(event.User),
        },
        Timestamp: time.Unix(0, int64(event.TimeStamp*1e9)),
        ThreadID:  event.ThreadTimeStamp,
        Mentions:  s.extractMentions(event.Text),
    }, nil
}

func (s *Service) HandleEventCallback(payload []byte) error {
    var callback slack.EventsAPIEvent
    if err := json.Unmarshal(payload, &callback); err != nil {
        return err
    }

    switch callback.Type {
    case slack.CallbackEvent:
        return s.processEvent(callback.InnerEvent)
    case slack.URLVerification:
        return s.handleVerification(callback)
    }

    return nil
}
```

### GitHub Integration

```go
package github

type Service struct {
    client  *github.Client
    oauth   *oauth.Provider
    webhook *webhook.Processor
}

func (s *Service) ProcessIssue(issue *github.Issue) (*normalizer.Message, error) {
    return &normalizer.Message{
        ID:      fmt.Sprintf("issue-%d", issue.GetNumber()),
        Source:  "github",
        Type:    "issue",
        Content: issue.GetBody(),
        Author: normalizer.Person{
            ID:   issue.GetUser().GetLogin(),
            Name: issue.GetUser().GetName(),
        },
        Timestamp: issue.GetCreatedAt().Time,
        Metadata: map[string]string{
            "repository": issue.GetRepository().GetFullName(),
            "state":      issue.GetState(),
            "labels":     s.extractLabels(issue.Labels),
        },
    }, nil
}

func (s *Service) HandleWebhook(payload []byte, event string) error {
    switch event {
    case "issues":
        return s.handleIssueEvent(payload)
    case "pull_request":
        return s.handlePullRequestEvent(payload)
    case "push":
        return s.handlePushEvent(payload)
    }

    return fmt.Errorf("unsupported event type: %s", event)
}
```

## API Endpoints

### Common Endpoints

Each integration service exposes these endpoints:

```
GET    /health                    # Health check
POST   /oauth/authorize          # Start OAuth flow
POST   /oauth/callback           # OAuth callback
POST   /webhook                  # Webhook endpoint
GET    /sync                     # Manual sync trigger
GET    /status                   # Integration status
```

### Service-Specific Endpoints

#### Gmail Service
```
GET    /api/v1/emails            # List recent emails
GET    /api/v1/emails/:id        # Get specific email
POST   /api/v1/sync/emails       # Sync email data
```

#### Slack Service
```
GET    /api/v1/channels          # List channels
GET    /api/v1/messages          # List messages
POST   /api/v1/sync/messages     # Sync message data
```

#### GitHub Service
```
GET    /api/v1/repositories      # List repositories
GET    /api/v1/issues            # List issues
GET    /api/v1/commits           # List commits
POST   /api/v1/sync/activity     # Sync repository activity
```

## Configuration

### Environment Variables

```bash
# Service Configuration
SERVICE_NAME=gmail-integration
SERVICE_PORT=8001
LOG_LEVEL=info

# Database
POSTGRES_URL=postgresql://user:pass@localhost/self_integrations

# Message Queue
NATS_URL=nats://localhost:4222

# OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8001/oauth/callback

# Webhook Configuration
WEBHOOK_SECRET=your-webhook-secret

# Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_DURATION=1h
```

### Service Configuration

```yaml
# configs/config.yaml
service:
  name: gmail-integration
  port: 8001
  timeout: 30s

oauth:
  provider: google
  client_id: ${GOOGLE_CLIENT_ID}
  client_secret: ${GOOGLE_CLIENT_SECRET}
  redirect_url: ${GOOGLE_REDIRECT_URL}
  scopes:
    - https://www.googleapis.com/auth/gmail.readonly
    - https://www.googleapis.com/auth/gmail.modify

webhook:
  path: /webhook
  secret: ${WEBHOOK_SECRET}
  timeout: 10s

database:
  url: ${POSTGRES_URL}
  max_connections: 10

nats:
  url: ${NATS_URL}
  subjects:
    - self.integrations.gmail
    - self.events.email

rate_limit:
  requests: 1000
  duration: 1h
```

## Development

### Running Locally

```bash
# Start infrastructure
docker-compose up -d postgres nats redis

# Start specific integration
cd integrations/google
go run cmd/main.go

# Start all integrations
docker-compose up
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Load tests
go test -tags=load ./...
```

### Adding New Integration

```bash
# Create service from template
./scripts/create-integration.sh new-service

# Implement OAuth flow
# Implement webhook handler
# Add data normalization
# Add API endpoints
# Update documentation
```

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
CMD ["./main"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gmail-integration
spec:
  replicas: 2
  selector:
    matchLabels:
      app: gmail-integration
  template:
    metadata:
      labels:
        app: gmail-integration
    spec:
      containers:
      - name: gmail-integration
        image: self/gmail-integration:latest
        ports:
        - containerPort: 8001
        env:
        - name: GOOGLE_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: google-credentials
              key: client_id
```

## Security

### Authentication
- OAuth 2.0 with PKCE for public clients
- Secure token storage with encryption
- Regular token rotation and refresh

### Webhook Security
- HMAC signature verification
- Request timestamp validation
- Rate limiting and DDoS protection

### Data Protection
- Encryption at rest and in transit
- Minimal data collection and retention
- User consent and data deletion compliance