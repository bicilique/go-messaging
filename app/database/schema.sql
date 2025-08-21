-- Create database (run this manually)
-- CREATE DATABASE go_messaging;

-- Users table to store user information
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_user_id BIGINT NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    language_code VARCHAR(10),
    is_bot BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create unique index for telegram_user_id (GORM compatible)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_user_id ON users(telegram_user_id);

-- Notification types table (static reference data)
CREATE TABLE IF NOT EXISTS notification_types (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    default_interval_minutes INTEGER DEFAULT 60,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create unique index for notification_types.code (GORM compatible)
CREATE UNIQUE INDEX IF NOT EXISTS idx_notification_types_code ON notification_types(code);

-- Subscriptions table to store user subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chat_id BIGINT NOT NULL,
    notification_type_id INTEGER NOT NULL REFERENCES notification_types(id),
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Subscription preferences (JSON for flexibility)
    preferences JSONB DEFAULT '{}',
    
    -- Tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_notified_at TIMESTAMP WITH TIME ZONE,
    
    -- Ensure unique subscription per user per type
    UNIQUE(user_id, notification_type_id)
);

-- Notification logs table (optional - for tracking sent notifications)
CREATE TABLE IF NOT EXISTS notification_logs (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'sent', -- sent, failed, delivered
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    error_message TEXT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_notification_type ON subscriptions(notification_type_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(is_active);
CREATE INDEX IF NOT EXISTS idx_subscriptions_chat_id ON subscriptions(chat_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_subscription_id ON notification_logs(subscription_id);
CREATE INDEX IF NOT EXISTS idx_notification_logs_sent_at ON notification_logs(sent_at);

-- Insert default notification types
INSERT INTO notification_types (code, name, description, default_interval_minutes) VALUES
('coinbase', 'Coinbase Alerts', 'Cryptocurrency price updates and market alerts', 1),
('news', 'News Alerts', 'Breaking news and important updates', 2),
('weather', 'Weather Updates', 'Weather forecasts and alerts', 4),
('price_alert', 'Price Alerts', 'Custom price threshold notifications', 5),
('custom', 'Custom Notifications', 'Custom notifications for specific needs', 6)
ON CONFLICT (code) DO NOTHING;

-- Update triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_types_updated_at BEFORE UPDATE ON notification_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
