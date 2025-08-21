# 🤖 Go Messaging Bot

A robust Telegram messaging bot built with Go, featuring rate limiting, message validation, and HTTP API endpoints for integration with external services.

## ✨ Features

- **🤖 Telegram Bot Integration**: Full-featured bot with command handling and message echoing
- **⏱️ Rate Limiting**: Prevents spam with configurable limits (10 messages/minute, 1 second intervals)
- **📏 Message Validation**: Enforces message length limits for better user experience
- **🌐 HTTP API**: RESTful endpoints for sending messages and notifications
- **🔄 Graceful Shutdown**: Context-aware polling with fast shutdown capabilities
- **📊 Structured Logging**: Comprehensive logging with user tracking and error monitoring
- **🐳 Docker Support**: Containerized deployment with Docker Compose
- **🏗️ Clean Architecture**: Modular design with separation of concerns

## 🚀 Quick Start

### Prerequisites

- Go 1.22.2 or higher
- Telegram Bot Token (get from [@BotFather](https://t.me/BotFather))
- Optional: Docker and Docker Compose

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/go-messaging.git
cd go-messaging
```

### 2. Environment Setup

Create a `.env` file in the `app` directory:

```env
# Required
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here

# Optional
PORT=8080
MODE=debug
DEVELOPER_HOST=true
```

### 3. Install Dependencies

```bash
cd app
go mod download
```

### 4. Build and Run

```bash
# Build the application
go build -o go-messaging ./cmd/main.go

# Run the bot
./go-messaging
```

### 5. Docker Deployment (Alternative)

```bash
# Using Docker Compose
docker-compose up -d
```

## 📋 Bot Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message and command overview |
| `/help` | Show available commands |
| `/limits` | Display current message and rate limits |
| `/status` | Check bot operational status |
| `/info` | Bot version and feature information |

## 🌐 HTTP API Endpoints

### Send Message
```http
POST /iris/send-message
Content-Type: application/json

{
  "chat_id": "123456789",
  "message": "Hello, World!"
}
```

## ⚙️ Configuration

### Message Limits
- **Regular Messages**: 1,000 characters
- **Commands**: 256 characters
- **Telegram Maximum**: 4,096 characters

### Rate Limiting
- **Messages per Minute**: 10
- **Minimum Interval**: 1 second between messages
- **Cleanup Interval**: 5 minutes (removes inactive users after 1 hour)

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | ✅ | - | Bot token from @BotFather |
| `TELEGRAM_CHAT_ID` | ✅ | - | Default chat ID for notifications |
| `PORT` | ❌ | 8080 | HTTP server port |
| `MODE` | ❌ | release | Gin mode (debug/release) |
| `DEVELOPER_HOST` | ❌ | false | Load .env file in development |

## 🏗️ Project Structure

```
go-messaging/
├── app/
│   ├── cmd/
│   │   └── main.go              # Application entry point
│   ├── config/
│   │   └── configurations.go    # Configuration management
│   ├── delivery/
│   │   └── http/
│   │       ├── iris_handler.go  # HTTP handlers
│   │       └── router.go        # Route configuration
│   ├── model/
│   │   ├── IrisWebHook.go      # Webhook data models
│   │   ├── TelegramBot.go      # Telegram API models
│   │   └── TelegramMessageRequest.go
│   ├── service/
│   │   └── telegram_service.go  # Core bot logic
│   ├── util/
│   │   └── FormatHelper.go     # Message formatting utilities
│   ├── go.mod
│   └── go.sum
├── docker-compose.yaml         # Docker Compose configuration
├── Dockerfile                  # Container build instructions
├── LICENSE                     # MIT License
└── README.md                   # This file
```

## 🔧 Development

### Running in Development Mode

```bash
cd app
export DEVELOPER_HOST=true
go run ./cmd/main.go
```

### Building for Production

```bash
cd app
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-messaging ./cmd/main.go
```

### Testing Rate Limiting

1. Send 11 messages quickly to trigger rate limiting
2. Wait 1 minute for rate limit reset
3. Check logs for rate limiting events

### Testing Message Validation

1. Send a message longer than 1,000 characters
2. Send a command longer than 256 characters
3. Verify rejection messages and logging

## 📊 Monitoring and Logging

The bot provides structured logging with the following information:

- **User Activity**: Track message counts and user interactions
- **Rate Limiting**: Log violations with user details
- **Message Validation**: Track length violations
- **System Events**: Bot startup, shutdown, and errors

### Log Examples

```json
{
  "level": "INFO",
  "msg": "Starting Telegram bot polling...",
  "time": "2025-08-21T14:25:16Z"
}

{
  "level": "WARN", 
  "msg": "Rate limit exceeded",
  "userID": 123456789,
  "username": "john_doe",
  "time": "2025-08-21T14:25:20Z"
}
```

## 🐳 Docker Deployment

### Build Image

```bash
docker build -t afiffaizianur/go-messaging:latest .
```

### Run with Docker Compose

```yaml
version: '3.8'
services:
  go-messaging:
    image: afiffaizianur/go-messaging:latest
    ports:
      - "8080:8080"
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID}
    restart: unless-stopped
```

## 🛡️ Security Features

- **Rate Limiting**: Prevents abuse and spam
- **Message Validation**: Prevents oversized messages
- **Input Sanitization**: Safe handling of user input
- **Graceful Shutdown**: Prevents data loss during restarts
- **Error Handling**: Comprehensive error management

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/go-messaging/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/go-messaging/discussions)
- **Telegram**: Contact [@BotFather](https://t.me/BotFather) for bot-related questions

## 🔗 Related Projects

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Go Telegram Bot Library](https://github.com/go-telegram/bot)

---

**Made with ❤️ by [Afif Faizianur](https://github.com/afiffaizianur)**
