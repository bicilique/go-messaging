# Go Messaging Bot - Universal Notification System

A powerful, scalable messaging system built with Go, featuring Telegram bot integration, admin approval system, and comprehensive user management with button-based interactions.

## 🌟 Features

### Core Features
- **🤖 Telegram Bot Integration**: Full-featured bot with inline keyboards and button interactions
- **👑 Admin Approval System**: Role-based user management with approval workflow
- **🔐 HTTP Basic Authentication**: Secure API endpoints with database-backed authentication
- **📱 Button-Based UI**: User-friendly Telegram interface without typing commands
- **⏰ Auto-Cleanup**: Automatic removal of pending users after 6 hours
- **🛡️ Rate Limiting**: Prevent spam and abuse
- **📊 Real-time Statistics**: User and system statistics
- **🔄 RESTful API**: Complete HTTP API for external integrations

### User Experience
- **No Command Typing**: Interactive button-based menus
- **Status Tracking**: Real-time approval status updates
- **Multi-Role Support**: Users and admins with different capabilities
- **Responsive Design**: Clean, intuitive Telegram interface

### Admin Features
- **Telegram Admin Panel**: Manage users directly through Telegram
- **HTTP Admin API**: RESTful endpoints for admin operations
- **Bulk Operations**: Approve, reject, disable users efficiently
- **User Statistics**: Monitor system usage and user growth
- **Audit Trail**: Track admin actions with timestamps

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Telegram Bot  │────│  Go Application │────│   PostgreSQL    │
│    (Frontend)   │    │   (Backend)     │    │   (Database)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                       ┌───────────────┐
                       │  HTTP API     │
                       │ (REST Endpoints)│
                       └───────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 13+
- Telegram Bot Token (get from [@BotFather](https://t.me/botfather))

### 1. Environment Setup

Create `.env` file:
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=go_messaging
DB_SSLMODE=disable

# Telegram Bot
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Server
PORT=8080
```

### 2. Database Setup

```sql
-- Create database
CREATE DATABASE go_messaging;

-- Run the schema
\i app/database/schema.sql

-- Initialize with admin data
\i init_database.sql
```

**Important**: Update the `telegram_user_id` values in `init_database.sql` with your actual Telegram User ID.

### 3. Installation & Run

```bash
cd app
go mod download
go run cmd/main.go
```

### 4. Create First Admin

```bash
# Method 1: HTTP API (with basic auth: admin/admin123)
curl -u admin:admin123 -X POST http://localhost:8080/api/v1/admin/create \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_user_id": YOUR_TELEGRAM_USER_ID,
    "username": "your_username",
    "first_name": "Your Name"
  }'

# Method 2: Update init_database.sql with your Telegram User ID and re-run
```

## 📱 User Interface

### For Regular Users

**Start Registration:**
```
User: /start
Bot: 🤖 Welcome to Go Messaging Bot!
     
     You need to register to receive notifications.
     Click the button below to get started.
     
     [🚀 Start Registration]  [ℹ️ About]  [❓ Help]
```

**Approved User Menu:**
```
Bot: ✅ Welcome Back!
     
     👤 John Doe
     🎉 You're approved to receive notifications!
     
     [🔔 Notifications]  [⚙️ Settings]
     [📊 My Status]      [📱 Subscriptions]
     [ℹ️ About]          [❓ Help]
```

### For Admins

**Admin Panel:**
```
Admin: /admin
Bot: 🔧 Admin Panel
     
     Welcome to the admin panel. Choose an option below:
     
     [📋 Pending Users]  [✅ Approved Users]
     [📊 User Stats]     [🧹 Cleanup]
```

**Pending Users Management:**
```
Bot: 📋 Pending Users (3):
     
     👤 John Doe (@johndoe)
     📅 Joined: 2025-01-15 10:30
     🆔 ID: abc123...
     
     [✅ Approve]  [❌ Reject]
     
     👤 Jane Smith (@janesmith)
     📅 Joined: 2025-01-15 11:15
     🆔 ID: def456...
     
     [✅ Approve]  [❌ Reject]
```

## 🔌 API Endpoints

### User Management
```http
GET    /api/v1/users                    # List all users
GET    /api/v1/users/:id               # Get user by ID
GET    /api/v1/users/telegram/:id      # Get user by Telegram ID
POST   /api/v1/users                   # Create/update user
PUT    /api/v1/users/:id               # Update user
DELETE /api/v1/users/telegram/:id      # Delete user
```

### Admin Operations (🔐 Basic Auth Required)
```http
POST   /api/v1/admin/create                    # Create admin
GET    /api/v1/admin/users/pending             # Get pending users
GET    /api/v1/admin/users/approved            # Get approved users
POST   /api/v1/admin/users/:id/approve         # Approve user
POST   /api/v1/admin/users/:id/reject          # Reject user
POST   /api/v1/admin/users/:id/disable         # Disable user
POST   /api/v1/admin/users/:id/enable          # Enable user
GET    /api/v1/admin/stats                     # Get user statistics
POST   /api/v1/admin/cleanup                   # Cleanup old pending users
```

### Authentication
All admin endpoints require HTTP Basic Authentication:
- Username: `admin`
- Password: `admin123` (change in production!)

## 🐳 Docker Deployment

### Using Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Manual Docker Build
```bash
# Build image
docker build -t go-messaging .

# Run container
docker run -d \
  -p 8080:8080 \
  -e TELEGRAM_BOT_TOKEN=your_token \
  -e DB_HOST=your_db_host \
  --name go-messaging \
  go-messaging
```

## 🔧 Configuration

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `go_messaging` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token | - |
| `PORT` | HTTP server port | `8080` |

### Database Tables
- `users` - User information and approval status
- `notification_types` - Available notification categories
- `subscriptions` - User notification subscriptions
- `notification_logs` - Sent notification history
- `api_credentials` - HTTP API authentication
- `app_config` - System configuration

## 📚 Usage Examples

### Postman Collection
Import `postman_collection.json` for ready-to-use API requests.

### OpenAPI Documentation
View `openapi.yaml` for complete API specification.

### Common Workflows

**1. User Registration Flow:**
```
User clicks /start → Registers → Waits for approval → Gets notified when approved
```

**2. Admin Approval Flow:**
```
Admin uses /admin → Views pending users → Approves/rejects with buttons
```

**3. User Management:**
```
Admin views approved users → Can disable/enable as needed
```

## 🔒 Security Features

- **Role-based Access Control**: Users vs Admins
- **HTTP Basic Authentication**: Secure API access
- **Rate Limiting**: Prevent spam and abuse
- **Input Validation**: Sanitize all inputs
- **Auto-cleanup**: Remove stale data automatically
- **Audit Logging**: Track all admin actions

## 🚧 Development

### Project Structure
```
├── app/
│   ├── cmd/main.go              # Application entry point
│   ├── config/                  # Configuration management
│   ├── database/                # Database setup and migrations
│   ├── delivery/http/           # HTTP handlers and middleware
│   ├── entity/                  # Data models
│   ├── internal/scheduler/      # Background jobs
│   ├── model/                   # Telegram bot models
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic
│   └── util/                    # Utility functions
├── migrations/                  # Database migrations
├── init_database.sql           # Database initialization
├── openapi.yaml               # API documentation
├── postman_collection.json    # Postman collection
└── docker-compose.yaml        # Docker setup
```

### Adding New Features
1. Update database schema in `app/database/schema.sql`
2. Add/update entities in `app/entity/`
3. Implement repository methods in `app/repository/`
4. Add business logic in `app/service/`
5. Create HTTP handlers in `app/delivery/http/`
6. Update routes in `router.go`
7. Add Telegram bot commands/callbacks as needed

### Testing
```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## 🎯 Use Cases

This system is perfect for:
- **Company Internal Notifications**: Employee alerts and updates
- **Community Management**: Member approval and notifications
- **Event Notifications**: Conference or meetup updates
- **System Monitoring**: Alert administrators about system events
- **Customer Support**: Managed notification system for customers
- **Educational Platforms**: Student and teacher notifications

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Contribution Guidelines
- Follow Go best practices
- Add tests for new features
- Update documentation
- Use conventional commit messages
- Ensure backwards compatibility

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check the `ADMIN_SYSTEM_README.md` for detailed admin features
- **Issues**: Report bugs on GitHub Issues
- **API Docs**: Use the OpenAPI specification in `openapi.yaml`
- **Postman**: Import the collection for API testing

## 🗺️ Roadmap

- [ ] **Multi-language Support**: i18n for different languages
- [ ] **Webhook Support**: Generic webhook notifications
- [ ] **Template System**: Customizable message templates
- [ ] **Scheduled Notifications**: Cron-based message scheduling
- [ ] **Analytics Dashboard**: Web-based admin dashboard
- [ ] **Message Threading**: Conversation management
- [ ] **File Attachments**: Support for images and documents
- [ ] **Custom Notification Types**: User-defined notification categories

---

**Built with ❤️ in Go** | **Ready for Production** | **Fully Documented** | **Docker Ready**
