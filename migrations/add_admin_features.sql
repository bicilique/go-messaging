-- Migration: Add admin and approval features
-- Run this migration to add admin functionality to existing database

-- Add new columns to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user',
ADD COLUMN IF NOT EXISTS approval_status VARCHAR(20) DEFAULT 'pending',
ADD COLUMN IF NOT EXISTS approved_by UUID,
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE;

-- Add foreign key constraint for approved_by
ALTER TABLE users 
ADD CONSTRAINT fk_users_approved_by 
FOREIGN KEY (approved_by) REFERENCES users(id);

-- Create indexes for new fields
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_approval_status ON users(approval_status);
CREATE INDEX IF NOT EXISTS idx_users_approved_by ON users(approved_by);

-- Create function to cleanup pending users older than 6 hours
CREATE OR REPLACE FUNCTION cleanup_pending_users()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM users 
    WHERE approval_status = 'pending' 
    AND created_at < NOW() - INTERVAL '6 hours';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Update existing users to be approved (for backward compatibility)
UPDATE users 
SET approval_status = 'approved',
    approved_at = NOW()
WHERE approval_status = 'pending';
