# 🤖 Go Messaging Bot - Smart Notification System

> **A modern, extensible Telegram bot that delivers personalized notifications** 
> 
> Built with Go + PostgreSQL • Clean Architecture • Production Ready

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-12+-336791?style=flat&logo=postgresql&logoColor=white)](https://postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 🌟 What This Bot Can Do

- 💰 **Crypto Price Alerts** - Get real-time cryptocurrency updates (BTC, ETH, ADA, DOT)
- 📰 **News Notifications** - Stay updated with breaking news (filtered by keywords)
- 🌤️ **Weather Updates** - Receive weather forecasts and alerts (location-based)
- 🚨 **Price Alerts** - Custom threshold notifications with configurable currency and threshold
- 🔔 **Custom Notifications** - Create your own alert types with custom messages

**Features:**
- ✅ **Parallel Scheduling** - All notification types run independently
- ✅ **Rate Limiting** - Prevents spam and abuse
- ✅ **User Management** - Automatic user creation and preference storage
- ✅ **Subscription Management** - Easy subscribe/unsubscribe with `/list` command
- ✅ **Development Mode** - Frequent notifications for testing (current)
- ✅ **Production Ready** - Easy switch to production intervals

**Simply text `/subscribe coinbase` and start receiving crypto updates!**

---

## 🚀 Quick Setup (5 Minutes)

### Step 1: Get Your Bot Ready
1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Create a new bot with `/newbot`
3. Copy your bot token (looks like `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)

### Step 2: Setup Database
```sql
-- Create PostgreSQL database
CREATE DATABASE go_messaging;
CREATE USER bot_user WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE go_messaging TO bot_user;
```

### Step 3: Clone & Configure
```bash
# Download the project
git clone https://github.com/your-repo/go-messaging.git
cd go-messaging/app

# Setup environment
cp .env.example .env
# Edit .env with your bot token and database info
```

### Step 4: Run!
```bash
go run ./cmd
```

🎉 **That's it!** Your bot is now live and ready to accept subscriptions!

---

## 🏗️ How It Works

This bot uses **Clean Architecture** for maximum maintainability:

```
📱 Telegram Bot → 🧠 Business Logic → 💾 PostgreSQL Database
```

**Simple Flow:**
1. User sends `/subscribe news`
2. Bot saves subscription to database
3. Bot automatically sends news updates
4. User can `/unsubscribe` anytime

---

## 🎯 Bot Commands (Super Simple!)

| Command | What It Does | Example |
|---------|--------------|---------|
| `/start` | 👋 Welcome message and setup | Just type `/start` |
| `/help` | ❓ Get help and command list | When you're stuck |
| `/types` | 📋 See all notification types | Shows: coinbase, news, weather... |
| `/subscribe <type>` | ✅ Start getting notifications | `/subscribe coinbase` |
| `/unsubscribe <type>` | ❌ Stop notifications | `/unsubscribe coinbase` |
| `/list` | 📄 Show your subscriptions | See what you're subscribed to |

### ✅ Current Status
- **All commands working** ✅
- **Subscription system functional** ✅  
- **Parallel notification scheduling** ✅
- **Rate limiting implemented** ✅
- **Database integration complete** ✅

### � Quick Examples
```
User: /subscribe price_alert
Bot:  ✅ Successfully subscribed to Price Alerts notifications!
      
      Default settings:
      • Currency: BTC
      • Threshold: $50,000
      • Interval: 5 minutes

User: /list
Bot:  � Your Active Subscriptions:
      
      🟢 Price Alerts - 5 min
      🟢 Coinbase Alerts - 1 min

User: /unsubscribe price_alert
Bot:  ✅ Unsubscribed from news alerts
```

---

## ⚙️ Configuration Made Easy

Create a `.env` file in the `app` folder:

```bash
# 🤖 Your Telegram Bot
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11

# 💾 Database Connection  
DB_HOST=localhost
DB_PORT=5432
DB_USER=bot_user
DB_PASSWORD=your_secure_password
DB_NAME=go_messaging

# 🔧 App Settings
PORT=8080
MODE=debug
DEVELOPER_HOST=true
```

> **💡 Pro Tip:** The bot automatically creates all database tables and sets up notification types on first run!

---

## 🔔 Available Notification Types

| Type | Code | What You Get | Development Schedule | Production Schedule |
|------|------|--------------|---------------------|---------------------|
| 💰 **Crypto Prices** | `coinbase` | Bitcoin, Ethereum prices | Every 1 minute | Every hour |
| 📰 **Breaking News** | `news` | Important news updates | Every 2 minutes | Every 30 min |
| 🌤️ **Weather** | `weather` | Weather forecasts | Every 4 minutes | Every 6 hours |
| 🚨 **Price Alerts** | `price_alert` | Custom price thresholds | Every 5 minutes | Every 5 min |
| 🔔 **Custom** | `custom` | Your custom notifications | Every 6 minutes | Every hour |

**Example:** Type `/subscribe coinbase` to get crypto updates!

### 🔧 Current Development Mode
The bot is currently configured for **development/testing** with frequent notifications:
- 🪙 **Coinbase**: Every 1 minute  
- 📰 **News**: Every 2 minutes
- 🌤️ **Weather**: Every 4 minutes
- 🚨 **Price Alert**: Every 5 minutes
- 🔔 **Custom**: Every 6 minutes

> **Note:** Price alerts in development mode always send notifications regardless of threshold conditions for testing purposes.

---

## 🚀 Switching from Development to Production

The bot is currently configured for **development/testing** with frequent notifications. To switch to production:

1. **Update notification intervals** in `cmd/main.go`:
   ```go
   notificationSchedule := map[string]int{
       "coinbase":    60,  // Every hour (was 1 minute)
       "news":        30,  // Every 30 minutes (was 2 minutes)
       "weather":     360, // Every 6 hours (was 4 minutes)
       "price_alert": 5,   // Every 5 minutes (unchanged)
       "custom":      60,  // Every hour (was 6 minutes)
   }
   ```

2. **Enable threshold checking** in `notification_dispatch_service.go`:
   ```go
   // Uncomment this production code:
   if currentPrice >= threshold {
       return fmt.Sprintf("🚨 Price Alert: %s\n\nCurrent price: $%.2f\nThreshold: $%.2f\n\nAlert triggered at %s",
           currency, currentPrice, threshold, time.Now().Format("15:04 MST")), nil
   }
   return "", fmt.Errorf("price threshold not met")
   ```

3. **Remove development notifications** that always send regardless of conditions.

---

## �️ For Developers

### 📁 Project Structure (Clean & Organized)
```
app/
├── 🚀 cmd/main.go              # Start here - main application
├── ⚙️ config/                  # Configuration management
├── 💾 database/                # Database setup & migrations
├── 📦 entity/                  # Data models (User, Subscription, etc.)
├── 🔄 repository/              # Database operations
├── 🧠 service/                 # Business logic
├── 🤖 telegram_bot_service.go  # Bot commands & responses
└── 🔧 model/                   # Helpers (rate limiting, validation)
```

### 🎯 Want to Add a New Notification Type?

**Super Easy! Just 3 steps:**

1. **Add to database:** (runs automatically)
   ```sql
   INSERT INTO notification_types (code, name, description, default_interval_minutes) 
   VALUES ('stocks', 'Stock Alerts', 'Stock price updates', 60);
   ```

2. **Add content generator:** (in `notification_dispatch_service.go`)
   ```go
   case "stocks":
       return s.getStockContent(ctx, preferences)
   ```

3. **Add to scheduler:** (in `main.go`)
   ```go
   notificationSchedule := map[string]int{
       "stocks": 120, // Every 2 hours
       // ... existing types
   }
   ```

**That's it!** Users can now use `/subscribe stocks`

### 🧪 Testing
```bash
# Run all tests
go test ./...

# Test with race detection
go test -race ./...

# Build for production
go build -o bot ./cmd
```

---

## 🚀 Deployment Options

### 🐳 Docker (Recommended)

**Prerequisites**: Docker and Docker Compose installed

```bash
# 1. Clone repository
git clone https://github.com/your-username/go-messaging.git
cd go-messaging

# 2. Configure environment
cp .env.example .env
# Edit .env and add your TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID

# 3. Start services (PostgreSQL + Go app)
docker-compose up -d

# 4. Check logs
docker-compose logs -f go-messaging

# 5. Verify health
curl http://localhost:8080/health
```

**What's included:**
- ✅ PostgreSQL database with automatic schema setup
- ✅ Go messaging application 
- ✅ Health checks and restart policies
- ✅ Persistent data storage
- ✅ Network isolation

### 📦 Binary
```bash
# Build
go build -o go-messaging ./cmd

# Run
./go-messaging
```

### ☁️ Cloud Deploy
Works great on:
- **Heroku** (with Heroku Postgres)
- **AWS** (with RDS)
- **Google Cloud** (with Cloud SQL)
- **DigitalOcean** (with Managed Database)

---

## 🔐 Security & Performance

✅ **Rate Limiting** - Prevents spam  
✅ **Input Validation** - Sanitizes all messages  
✅ **SQL Injection Protection** - Uses prepared statements  
✅ **Graceful Shutdown** - Safe restarts  
✅ **Environment Variables** - No hardcoded secrets  
✅ **Comprehensive Logging** - Track everything  

---

## 💡 FAQ

<details>
<summary><strong>❓ How do I get a Telegram bot token?</strong></summary>

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` command
3. Follow the instructions to name your bot
4. Copy the token (format: `123456:ABC-DEF1234...`)
5. Paste it in your `.env` file

</details>

<details>
<summary><strong>❓ Can I add my own notification types?</strong></summary>

Yes! It's super easy:
1. Add your notification type to the database
2. Create a content generator function
3. Add it to the scheduler

See the "For Developers" section above for detailed steps.

</details>

<details>
<summary><strong>❓ How do I customize notification intervals?</strong></summary>

Edit the `main.go` file and modify the `notificationSchedule` map:
```go
notificationSchedule := map[string]int{
    "coinbase": 30,  // Every 30 minutes instead of 60
    "news":     15,  // Every 15 minutes instead of 30
}
```

</details>

<details>
<summary><strong>❓ Is this production ready?</strong></summary>

Yes! The bot includes:
- ✅ Rate limiting and spam protection
- ✅ Database connection pooling
- ✅ Graceful shutdown handling
- ✅ Comprehensive error logging
- ✅ Input validation and sanitization

</details>

---

## 🤝 Contributing

We love contributions! Here's how to help:

1. **🍴 Fork** the repository
2. **🌟 Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **💾 Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **🚀 Push** to the branch (`git push origin feature/amazing-feature`)
5. **📝 Open** a Pull Request

### 🎯 Ideas for Contributions
- Add new notification types (stocks, sports, etc.)
- Integrate real APIs (replace mock data)
- Add user preference management
- Create admin dashboard
- Add Docker deployment scripts

---

## � License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Credits & Thanks

Built with these amazing tools:

- 🤖 [go-telegram/bot](https://github.com/go-telegram/bot) - Telegram Bot API
- 🗄️ [GORM](https://gorm.io/) - Go ORM library  
- 🐘 [PostgreSQL](https://www.postgresql.org/) - Database system
- 🚀 [Go](https://golang.org/) - Programming language

---

## 🎉 Ready to Start?

```bash
git clone https://github.com/your-repo/go-messaging.git
cd go-messaging/app
cp .env.example .env
# Edit .env with your bot token
go run ./cmd
```

**Your notification bot is now live! 🚀**

Need help? Open an issue or check our [documentation](docs/) for more details.
