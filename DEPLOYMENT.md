# 🚀 Go Messaging Bot - Deployment Guide

This guide walks you through deploying the Go Messaging Bot in different environments.

## 📋 Pre-Deployment Checklist

- [ ] Telegram Bot Token from [@BotFather](https://t.me/botfather)
- [ ] PostgreSQL database (local or cloud)
- [ ] Your Telegram User ID (use `./get_telegram_id.sh`)
- [ ] Go 1.21+ (for source deployment)
- [ ] Docker & Docker Compose (for container deployment)

## 🐳 Docker Deployment (Recommended)

### Quick Start with Docker Compose

1. **Clone and Configure**
```bash
git clone <your-repo>
cd go-messaging
cp .env.example .env
```

2. **Edit Environment Variables**
```bash
# Edit .env file
nano .env
```

Required variables:
```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=go_messaging
DB_SSLMODE=disable

# Telegram
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11

# Server
PORT=8080
```

3. **Update Admin Configuration**
```bash
# Edit with your Telegram User ID
nano init_database.sql

# Replace 123456789 with your actual Telegram User ID
# You can add multiple admins here
```

4. **Deploy**
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Check status
docker-compose ps
```

5. **Initialize Database** (first time only)
```bash
# Wait for postgres to be ready, then:
docker-compose exec postgres psql -U postgres -d go_messaging -f /docker-entrypoint-initdb.d/init_database.sql
```

6. **Test Your Bot**
- Message your bot on Telegram: `/start`
- Use admin panel: `/admin`
- Test API: `curl http://localhost:8080/api/v1/admin/stats -u admin:admin123`

### Manual Docker Build

```bash
# Build image
docker build -t go-messaging .

# Run with external database
docker run -d \
  --name go-messaging \
  -p 8080:8080 \
  -e TELEGRAM_BOT_TOKEN=your_token \
  -e DB_HOST=your_db_host \
  -e DB_USER=your_db_user \
  -e DB_PASSWORD=your_db_password \
  -e DB_NAME=go_messaging \
  go-messaging

# Check logs
docker logs -f go-messaging
```

## 🖥️ Source Code Deployment

### Local Development

1. **Prerequisites**
```bash
# Install Go 1.21+
go version

# Install PostgreSQL
# macOS: brew install postgresql
# Ubuntu: sudo apt install postgresql postgresql-contrib
```

2. **Database Setup**
```sql
-- Connect to PostgreSQL
psql -U postgres

-- Create database and user
CREATE DATABASE go_messaging;
CREATE USER bot_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE go_messaging TO bot_user;
\q

-- Run schema
psql -U bot_user -d go_messaging -f app/database/schema.sql

-- Initialize data (edit Telegram User ID first!)
psql -U bot_user -d go_messaging -f init_database.sql
```

3. **Application Setup**
```bash
cd app

# Install dependencies
go mod download

# Set environment variables
export TELEGRAM_BOT_TOKEN="your_bot_token"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="bot_user"
export DB_PASSWORD="secure_password"
export DB_NAME="go_messaging"
export DB_SSLMODE="disable"

# Run application
go run cmd/main.go
```

### Production Server (Linux)

1. **System Setup**
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Go
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib -y
```

2. **Database Setup**
```bash
sudo -u postgres psql
CREATE DATABASE go_messaging;
CREATE USER bot_user WITH PASSWORD 'very_secure_password';
GRANT ALL PRIVILEGES ON DATABASE go_messaging TO bot_user;
\q
```

3. **Application Deployment**
```bash
# Clone repository
git clone <your-repo>
cd go-messaging/app

# Build application
go mod download
go build -o go-messaging cmd/main.go

# Create service user
sudo useradd -r -s /bin/false gomessaging

# Create directories
sudo mkdir -p /opt/gomessaging
sudo mkdir -p /var/log/gomessaging

# Copy files
sudo cp go-messaging /opt/gomessaging/
sudo cp -r templates /opt/gomessaging/ (if any)
sudo chown -R gomessaging:gomessaging /opt/gomessaging
sudo chown -R gomessaging:gomessaging /var/log/gomessaging
```

4. **Systemd Service**
```bash
# Create service file
sudo nano /etc/systemd/system/gomessaging.service
```

```ini
[Unit]
Description=Go Messaging Bot
After=network.target postgresql.service

[Service]
Type=simple
User=gomessaging
Group=gomessaging
WorkingDirectory=/opt/gomessaging
ExecStart=/opt/gomessaging/go-messaging
Restart=always
RestartSec=5

# Environment variables
Environment=TELEGRAM_BOT_TOKEN=your_bot_token
Environment=DB_HOST=localhost
Environment=DB_PORT=5432
Environment=DB_USER=bot_user
Environment=DB_PASSWORD=very_secure_password
Environment=DB_NAME=go_messaging
Environment=DB_SSLMODE=disable
Environment=PORT=8080

# Logging
StandardOutput=append:/var/log/gomessaging/app.log
StandardError=append:/var/log/gomessaging/error.log

[Install]
WantedBy=multi-user.target
```

5. **Start Service**
```bash
# Reload systemd and start service
sudo systemctl daemon-reload
sudo systemctl enable gomessaging
sudo systemctl start gomessaging

# Check status
sudo systemctl status gomessaging

# View logs
sudo journalctl -u gomessaging -f
```

## ☁️ Cloud Deployment

### Heroku

1. **Prepare for Heroku**
```bash
# Create Procfile
echo "web: ./main" > Procfile

# Create heroku.yml for container deployment
cat > heroku.yml << EOF
build:
  docker:
    web: Dockerfile
EOF
```

2. **Deploy to Heroku**
```bash
# Install Heroku CLI and login
heroku login

# Create app
heroku create your-messaging-bot

# Add PostgreSQL addon
heroku addons:create heroku-postgresql:hobby-dev

# Set environment variables
heroku config:set TELEGRAM_BOT_TOKEN=your_token

# Set stack to container
heroku stack:set container

# Deploy
git push heroku main

# Initialize database
heroku pg:psql < init_database.sql
```

### DigitalOcean App Platform

1. **Create app.yaml**
```yaml
name: go-messaging-bot
services:
- name: web
  source_dir: /
  github:
    repo: your-username/go-messaging
    branch: main
  run_command: ./main
  environment_slug: go
  http_port: 8080
  instance_count: 1
  instance_size_slug: basic-xxs
  envs:
  - key: TELEGRAM_BOT_TOKEN
    value: your_bot_token
  - key: DB_HOST
    value: ${db.HOSTNAME}
  - key: DB_PORT
    value: ${db.PORT}
  - key: DB_USER
    value: ${db.USERNAME}
  - key: DB_PASSWORD
    value: ${db.PASSWORD}
  - key: DB_NAME
    value: ${db.DATABASE}
databases:
- name: db
  engine: PG
  num_nodes: 1
  size: db-s-dev-database
```

### AWS ECS/Fargate

1. **Build and Push to ECR**
```bash
# Create ECR repository
aws ecr create-repository --repository-name go-messaging

# Get login token
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build and tag
docker build -t go-messaging .
docker tag go-messaging:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/go-messaging:latest

# Push
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/go-messaging:latest
```

2. **Create ECS Task Definition** (use AWS Console or CLI)

3. **Create RDS PostgreSQL Database** (use AWS Console)

## 🔧 Post-Deployment Configuration

### 1. Verify Deployment
```bash
# Check API health
curl http://your-domain:8080/api/v1/admin/stats -u admin:admin123

# Check logs
# Docker: docker-compose logs -f app
# Systemd: sudo journalctl -u gomessaging -f
```

### 2. Create First Admin
```bash
# Via API
curl -u admin:admin123 -X POST http://your-domain:8080/api/v1/admin/create \
  -H "Content-Type: application/json" \
  -d '{
    "telegram_user_id": YOUR_TELEGRAM_USER_ID,
    "username": "your_username",
    "first_name": "Your Name"
  }'

# Via Database (if API doesn't work)
psql -d go_messaging -c "INSERT INTO users (telegram_user_id, username, first_name, role, approval_status, approved_at) VALUES (YOUR_USER_ID, 'admin', 'Admin', 'admin', 'approved', NOW());"
```

### 3. Test Bot Functionality
1. Message your bot: `/start`
2. Use admin panel: `/admin`
3. Try user registration flow
4. Test API endpoints

### 4. Security Hardening

**Change Default Credentials:**
```sql
-- Update API credentials
UPDATE api_credentials SET password_hash = '$2a$10$new_hash_here' WHERE username = 'admin';
```

**Generate new hash:**
```bash
# Using bcrypt online tool or:
go run -c "package main; import(\"golang.org/x/crypto/bcrypt\"; \"fmt\"); func main(){h,_:=bcrypt.GenerateFromPassword([]byte(\"your_new_password\"),10);fmt.Println(string(h))}"
```

**Firewall Rules:**
```bash
# Only allow necessary ports
sudo ufw allow 22    # SSH
sudo ufw allow 8080  # App (or 80/443 if using reverse proxy)
sudo ufw allow 5432  # PostgreSQL (only if external access needed)
sudo ufw enable
```

### 5. Monitoring & Logging

**Set up log rotation:**
```bash
# Create logrotate config
sudo nano /etc/logrotate.d/gomessaging

/var/log/gomessaging/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    postrotate
        systemctl reload gomessaging
    endscript
}
```

**Basic monitoring script:**
```bash
#!/bin/bash
# health_check.sh
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/admin/stats -u admin:admin123)
if [ $response != "200" ]; then
    echo "App is down! HTTP: $response"
    # Restart service or send alert
    systemctl restart gomessaging
fi
```

## 🔍 Troubleshooting

### Common Issues

**1. Database Connection Issues**
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Check connection
pg_isready -h localhost -p 5432

# Test connection with credentials
psql -h localhost -p 5432 -U bot_user -d go_messaging -c "SELECT version();"
```

**2. Telegram Bot Not Responding**
```bash
# Verify bot token
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe

# Check webhook status
curl https://api.telegram.org/bot<YOUR_TOKEN>/getWebhookInfo

# If webhook is set, delete it (we use polling)
curl https://api.telegram.org/bot<YOUR_TOKEN>/deleteWebhook
```

**3. Application Won't Start**
```bash
# Check logs
journalctl -u gomessaging -n 50

# Check environment variables
systemctl show gomessaging --property=Environment

# Test manual start
sudo -u gomessaging /opt/gomessaging/go-messaging
```

**4. Admin Panel Not Working**
```bash
# Verify admin user exists
psql -d go_messaging -c "SELECT * FROM users WHERE role='admin';"

# Check if user has correct Telegram ID
# Use @userinfobot to get your ID

# Manually create admin if needed
psql -d go_messaging -c "INSERT INTO users (telegram_user_id, username, first_name, role, approval_status, approved_at) VALUES (YOUR_ID, 'admin', 'Admin', 'admin', 'approved', NOW());"
```

### Getting Help

1. **Check logs first** - Most issues are evident in logs
2. **Verify configuration** - Double-check environment variables
3. **Test components individually** - Database, Telegram API, HTTP endpoints
4. **Check firewall/network** - Ensure ports are accessible
5. **Review documentation** - README.md and ADMIN_SYSTEM_README.md

### Performance Tuning

**For high-volume deployments:**
- Increase PostgreSQL connection pool
- Use Redis for caching
- Set up horizontal scaling with load balancer
- Monitor memory and CPU usage
- Consider rate limiting adjustments

---

**🎉 Congratulations!** Your Go Messaging Bot should now be deployed and ready to use!

**Next Steps:**
1. Share your bot with users
2. Monitor usage and performance
3. Customize notification types for your use case
4. Set up monitoring and alerting
5. Plan for scaling as your user base grows
