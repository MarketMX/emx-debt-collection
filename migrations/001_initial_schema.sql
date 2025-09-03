-- Initial schema for Age Analysis Messaging Application (AAMA)
-- Migration: 001_initial_schema
-- Created: 2025-09-01
-- Description: Creates the core tables for user management, uploads, accounts, and message logs

-- Extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table (for Keycloak integration - stores user references and metadata)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    keycloak_id VARCHAR(255) UNIQUE NOT NULL, -- Keycloak user ID for reference
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Uploads table - tracks file uploads and their processing status
CREATE TABLE uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(500) NOT NULL,
    original_filename VARCHAR(500) NOT NULL,
    file_path VARCHAR(1000), -- Stored file path
    file_size BIGINT, -- File size in bytes
    mime_type VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    processing_started_at TIMESTAMP WITH TIME ZONE,
    processing_completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT, -- Store any processing errors
    records_processed INTEGER DEFAULT 0,
    records_failed INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Accounts table - stores parsed account data from uploads
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    upload_id UUID NOT NULL REFERENCES uploads(id) ON DELETE CASCADE,
    account_code VARCHAR(100) NOT NULL, -- Maps to "Account" column in Excel
    customer_name VARCHAR(500) NOT NULL, -- Maps to "Name" column in Excel
    contact_person VARCHAR(500), -- Maps to "Contact" column in Excel (nullable)
    telephone VARCHAR(50) NOT NULL, -- Maps to "Telephone" column in Excel
    amount_current DECIMAL(15,2) DEFAULT 0.00, -- Maps to "Current" column
    amount_30d DECIMAL(15,2) DEFAULT 0.00, -- Maps to "30 Days" column
    amount_60d DECIMAL(15,2) DEFAULT 0.00, -- Maps to "60 Days" column
    amount_90d DECIMAL(15,2) DEFAULT 0.00, -- Maps to "90 Days" column
    amount_120d DECIMAL(15,2) DEFAULT 0.00, -- Maps to "120 Days" column
    total_balance DECIMAL(15,2) NOT NULL, -- Maps to "Total Balance" column
    is_selected BOOLEAN DEFAULT false, -- For UI selection state
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Message logs table - tracks all messaging attempts
CREATE TABLE message_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    upload_id UUID NOT NULL REFERENCES uploads(id) ON DELETE CASCADE, -- Denormalized for easier querying
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Who triggered the message
    message_template TEXT NOT NULL, -- The message template used
    message_content TEXT NOT NULL, -- The actual message sent (after template substitution)
    recipient_telephone VARCHAR(50) NOT NULL, -- Phone number message was sent to
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'delivered', 'read')),
    external_message_id VARCHAR(255), -- ID from external messaging service
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    response_from_service TEXT, -- Raw response from messaging service
    error_message TEXT, -- Any error details
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance

-- Users indexes
CREATE INDEX idx_users_keycloak_id ON users(keycloak_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Uploads indexes
CREATE INDEX idx_uploads_user_id ON uploads(user_id);
CREATE INDEX idx_uploads_status ON uploads(status);
CREATE INDEX idx_uploads_created_at ON uploads(created_at DESC);
CREATE INDEX idx_uploads_user_status ON uploads(user_id, status);

-- Accounts indexes
CREATE INDEX idx_accounts_upload_id ON accounts(upload_id);
CREATE INDEX idx_accounts_account_code ON accounts(account_code);
CREATE INDEX idx_accounts_customer_name ON accounts(customer_name);
CREATE INDEX idx_accounts_telephone ON accounts(telephone);
CREATE INDEX idx_accounts_total_balance ON accounts(total_balance DESC);
CREATE INDEX idx_accounts_is_selected ON accounts(is_selected);
-- Composite index for common queries
CREATE INDEX idx_accounts_upload_selected ON accounts(upload_id, is_selected);

-- Message logs indexes
CREATE INDEX idx_message_logs_account_id ON message_logs(account_id);
CREATE INDEX idx_message_logs_upload_id ON message_logs(upload_id);
CREATE INDEX idx_message_logs_user_id ON message_logs(user_id);
CREATE INDEX idx_message_logs_status ON message_logs(status);
CREATE INDEX idx_message_logs_sent_at ON message_logs(sent_at DESC);
CREATE INDEX idx_message_logs_recipient_telephone ON message_logs(recipient_telephone);
-- Composite indexes for common queries
CREATE INDEX idx_message_logs_upload_status ON message_logs(upload_id, status);
CREATE INDEX idx_message_logs_account_status ON message_logs(account_id, status);

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_uploads_updated_at BEFORE UPDATE ON uploads
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_message_logs_updated_at BEFORE UPDATE ON message_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add some useful views for common queries

-- View for upload summary with account counts
CREATE VIEW upload_summary AS
SELECT 
    u.id,
    u.filename,
    u.original_filename,
    u.status,
    u.created_at,
    u.user_id,
    usr.email as user_email,
    COUNT(a.id) as total_accounts,
    COUNT(CASE WHEN a.is_selected THEN 1 END) as selected_accounts,
    SUM(a.total_balance) as total_balance_sum,
    SUM(CASE WHEN a.is_selected THEN a.total_balance ELSE 0 END) as selected_balance_sum
FROM uploads u
LEFT JOIN accounts a ON u.id = a.upload_id
LEFT JOIN users usr ON u.user_id = usr.id
GROUP BY u.id, u.filename, u.original_filename, u.status, u.created_at, u.user_id, usr.email;

-- View for message log summary
CREATE VIEW message_log_summary AS
SELECT 
    ml.upload_id,
    COUNT(*) as total_messages,
    COUNT(CASE WHEN ml.status = 'sent' THEN 1 END) as sent_messages,
    COUNT(CASE WHEN ml.status = 'failed' THEN 1 END) as failed_messages,
    COUNT(CASE WHEN ml.status = 'delivered' THEN 1 END) as delivered_messages,
    MIN(ml.sent_at) as first_sent_at,
    MAX(ml.sent_at) as last_sent_at
FROM message_logs ml
GROUP BY ml.upload_id;

-- Add comments for documentation
COMMENT ON TABLE users IS 'User accounts integrated with Keycloak authentication';
COMMENT ON TABLE uploads IS 'File uploads and their processing status';
COMMENT ON TABLE accounts IS 'Parsed account data from uploaded age analysis reports';
COMMENT ON TABLE message_logs IS 'Log of all messaging attempts and their status';

COMMENT ON COLUMN users.keycloak_id IS 'Keycloak user ID for authentication integration';
COMMENT ON COLUMN uploads.status IS 'Processing status: pending, processing, completed, failed';
COMMENT ON COLUMN accounts.is_selected IS 'UI selection state for bulk operations';
COMMENT ON COLUMN message_logs.status IS 'Message status: pending, sent, failed, delivered, read';
COMMENT ON COLUMN message_logs.retry_count IS 'Number of retry attempts made';
COMMENT ON COLUMN message_logs.external_message_id IS 'ID returned from external messaging service';
