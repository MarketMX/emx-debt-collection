-- Migration: 002_add_engagemx_client_id
-- Created: 2025-09-10
-- Description: Add engagemx_client_id column to users table for multitenancy support

-- Add the engagemx_client_id column to users table
ALTER TABLE users 
ADD COLUMN engagemx_client_id VARCHAR(255) NOT NULL DEFAULT 'default';

-- Create index for performance on multitenancy queries
CREATE INDEX idx_users_engagemx_client_id ON users(engagemx_client_id);

-- Create composite index for common multitenancy queries
CREATE INDEX idx_users_client_active ON users(engagemx_client_id, is_active);

-- Add comment for documentation
COMMENT ON COLUMN users.engagemx_client_id IS 'Client identifier for multitenancy from MMx-Dashboard';

-- Update the upload_summary view to include client_id for multitenancy
DROP VIEW IF EXISTS upload_summary;
CREATE VIEW upload_summary AS
SELECT 
    u.id,
    u.filename,
    u.original_filename,
    u.status,
    u.created_at,
    u.user_id,
    usr.email as user_email,
    usr.engagemx_client_id,
    COUNT(a.id) as total_accounts,
    COUNT(CASE WHEN a.is_selected THEN 1 END) as selected_accounts,
    SUM(a.total_balance) as total_balance_sum,
    SUM(CASE WHEN a.is_selected THEN a.total_balance ELSE 0 END) as selected_balance_sum
FROM uploads u
LEFT JOIN accounts a ON u.id = a.upload_id
LEFT JOIN users usr ON u.user_id = usr.id
GROUP BY u.id, u.filename, u.original_filename, u.status, u.created_at, u.user_id, usr.email, usr.engagemx_client_id;

-- Update the message_log_summary view to include client_id for multitenancy
DROP VIEW IF EXISTS message_log_summary;
CREATE VIEW message_log_summary AS
SELECT 
    ml.upload_id,
    usr.engagemx_client_id,
    COUNT(*) as total_messages,
    COUNT(CASE WHEN ml.status = 'sent' THEN 1 END) as sent_messages,
    COUNT(CASE WHEN ml.status = 'failed' THEN 1 END) as failed_messages,
    COUNT(CASE WHEN ml.status = 'delivered' THEN 1 END) as delivered_messages,
    MIN(ml.sent_at) as first_sent_at,
    MAX(ml.sent_at) as last_sent_at
FROM message_logs ml
JOIN uploads u ON ml.upload_id = u.id
JOIN users usr ON ml.user_id = usr.id
GROUP BY ml.upload_id, usr.engagemx_client_id;