-- Database Initialization Script
-- This script creates initial admin users and required data for the Go Messaging Bot

-- ================================
-- API CREDENTIALS SETUP
-- ================================

-- Insert default admin API credentials (password: admin123)
-- In production, you should change these credentials immediately
INSERT INTO api_credentials (username, password_hash, role, is_active) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye6vKH.h.0KJ3f4.e7.e8Qs6S4K9Z6jWG', 'admin', true),
-- ('moderator', '$2a$10$N9qo8uLOickgx2ZMRZoMye6vKH.h.0KJ3f4.e7.e8Qs6S4K9Z6jWG', 'moderator', true)
ON CONFLICT (username) DO NOTHING;

-- ================================
-- TELEGRAM ADMIN USERS SETUP
-- ================================

-- Create initial admin users
-- Replace these telegram_user_id values with your actual Telegram IDs
-- You can get your Telegram ID by messaging @userinfobot on Telegram

-- Example admin user (REPLACE WITH YOUR ACTUAL TELEGRAM USER ID)
INSERT INTO users (
    telegram_user_id, 
    username, 
    first_name, 
    last_name, 
    role, 
    approval_status, 
    approved_at
) VALUES
-- IMPORTANT: Replace 123456789 with your actual Telegram User ID
(630499194, 'afif', 'afif', 'faizianur', 'admin', 'approved', NOW()),
ON CONFLICT (telegram_user_id) DO UPDATE SET
    role = EXCLUDED.role,
    approval_status = EXCLUDED.approval_status,
    approved_at = EXCLUDED.approved_at;
-- ================================
-- NOTIFICATION TYPES SETUP
-- ================================

-- Ensure notification types are properly set up
INSERT INTO notification_types (code, name, description, default_interval_minutes, is_active) VALUES
-- ('maintenance', 'Maintenance Alerts', 'Scheduled maintenance notifications', 60, true),
-- ('system', 'System Notifications', 'Important system alerts and updates', 1, true),
('security', 'Security Alerts', 'Security-related notifications', 9223372036854775807, true), 
('general', 'General Notifications', 'General purpose notifications', 9223372036854775807, true)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    default_interval_minutes = EXCLUDED.default_interval_minutes,
    is_active = EXCLUDED.is_active;

-- ================================
-- CONFIGURATION DATA
-- ================================

-- Create a configuration table for app settings
CREATE TABLE IF NOT EXISTS app_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(255) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default configuration
INSERT INTO app_config (config_key, config_value, description) VALUES
('cleanup_interval_hours', '6', 'Hours after which pending users are automatically cleaned up'),
('max_pending_users', '1000', 'Maximum number of pending users allowed'),
('welcome_message', 'Welcome to our notification bot! Please wait for admin approval.', 'Default welcome message for new users'),
('admin_notification_enabled', 'true', 'Whether to notify admins about new user registrations'),
('rate_limit_messages_per_minute', '10', 'Rate limit for bot messages per user per minute')
ON CONFLICT (config_key) DO NOTHING;

-- ================================
-- INDEXES AND PERFORMANCE
-- ================================

-- Additional indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_role_status ON users(role, approval_status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_chat_active ON subscriptions(chat_id, is_active);

-- ================================
-- FUNCTIONS AND TRIGGERS
-- ================================

-- Function to update app_config updated_at
CREATE OR REPLACE FUNCTION update_app_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for app_config
CREATE TRIGGER update_app_config_updated_at 
    BEFORE UPDATE ON app_config
    FOR EACH ROW 
    EXECUTE FUNCTION update_app_config_updated_at();

-- Function to get active admin count
CREATE OR REPLACE FUNCTION get_active_admin_count()
RETURNS INTEGER AS $$
BEGIN
    RETURN (SELECT COUNT(*) FROM users WHERE role = 'admin' AND approval_status = 'approved');
END;
$$ LANGUAGE plpgsql;

-- ================================
-- VERIFICATION QUERIES
-- ================================

-- Display setup results
DO $$
DECLARE
    admin_count INTEGER;
    api_cred_count INTEGER;
    notification_type_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO admin_count FROM users WHERE role = 'admin';
    SELECT COUNT(*) INTO api_cred_count FROM api_credentials WHERE is_active = true;
    SELECT COUNT(*) INTO notification_type_count FROM notification_types WHERE is_active = true;
    
    RAISE NOTICE 'Database initialization completed:';
    RAISE NOTICE '- Admin users created: %', admin_count;
    RAISE NOTICE '- API credentials created: %', api_cred_count;
    RAISE NOTICE '- Active notification types: %', notification_type_count;
    RAISE NOTICE '';
    RAISE NOTICE 'IMPORTANT: Update the telegram_user_id values in the users table with your actual Telegram User IDs';
    RAISE NOTICE 'IMPORTANT: Change the default API credentials in production';
END
$$;

-- Show created admin users (for verification)
SELECT 
    'Admin Users Created:' as info,
    telegram_user_id,
    username,
    first_name,
    role,
    approval_status,
    created_at
FROM users 
WHERE role = 'admin'
ORDER BY created_at;

-- Show API credentials (for verification)
SELECT 
    'API Credentials Created:' as info,
    username,
    role,
    is_active,
    created_at
FROM api_credentials
ORDER BY created_at;
