-- Migration script to convert from BIGSERIAL to UUID for users table
-- WARNING: This migration is destructive and will require rebuilding relationships
-- Make sure to backup your data before running this migration

-- Step 1: Create a backup of existing data
CREATE TABLE users_backup AS SELECT * FROM users;
CREATE TABLE subscriptions_backup AS SELECT * FROM subscriptions;

-- Step 2: Drop existing foreign key constraints
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_user_id_fkey;

-- Step 3: Add UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Step 4: Add a new UUID column to users table
ALTER TABLE users ADD COLUMN id_uuid UUID DEFAULT uuid_generate_v4();

-- Step 5: Update all existing users to have UUIDs
UPDATE users SET id_uuid = uuid_generate_v4() WHERE id_uuid IS NULL;

-- Step 6: Create a mapping table for migration
CREATE TABLE user_id_mapping AS 
SELECT id as old_id, id_uuid as new_id FROM users;

-- Step 7: Add new UUID column to subscriptions
ALTER TABLE subscriptions ADD COLUMN user_id_uuid UUID;

-- Step 8: Update subscriptions to use new UUIDs
UPDATE subscriptions 
SET user_id_uuid = (
    SELECT new_id 
    FROM user_id_mapping 
    WHERE old_id = subscriptions.user_id
);

-- Step 9: Drop old columns and rename new ones
ALTER TABLE users DROP COLUMN id CASCADE;
ALTER TABLE users RENAME COLUMN id_uuid TO id;
ALTER TABLE users ADD PRIMARY KEY (id);

ALTER TABLE subscriptions DROP COLUMN user_id;
ALTER TABLE subscriptions RENAME COLUMN user_id_uuid TO user_id;
ALTER TABLE subscriptions ALTER COLUMN user_id SET NOT NULL;

-- Step 10: Recreate foreign key constraints
ALTER TABLE subscriptions 
ADD CONSTRAINT subscriptions_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Step 11: Recreate indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_user_id ON users(telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_notification_type ON subscriptions(notification_type_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(is_active);
CREATE INDEX IF NOT EXISTS idx_subscriptions_chat_id ON subscriptions(chat_id);

-- Step 12: Clean up temporary tables
DROP TABLE user_id_mapping;

-- Step 13: Verify migration
SELECT 
    'Users count' as table_name, 
    COUNT(*) as record_count 
FROM users
UNION ALL
SELECT 
    'Subscriptions count' as table_name, 
    COUNT(*) as record_count 
FROM subscriptions
UNION ALL
SELECT 
    'Users backup count' as table_name, 
    COUNT(*) as record_count 
FROM users_backup
UNION ALL
SELECT 
    'Subscriptions backup count' as table_name, 
    COUNT(*) as record_count 
FROM subscriptions_backup;

-- If everything looks good, you can drop the backup tables:
-- DROP TABLE users_backup;
-- DROP TABLE subscriptions_backup;
