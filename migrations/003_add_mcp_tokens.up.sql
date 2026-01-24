-- Create mcp_tokens table for storing MCP authentication tokens
CREATE TABLE IF NOT EXISTS mcp_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token UUID UNIQUE NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Create index on token for fast lookup
CREATE INDEX IF NOT EXISTS idx_mcp_tokens_token ON mcp_tokens(token);

-- Create index on user_id for listing user's tokens
CREATE INDEX IF NOT EXISTS idx_mcp_tokens_user_id ON mcp_tokens(user_id);

-- Create index on is_active for filtering active tokens
CREATE INDEX IF NOT EXISTS idx_mcp_tokens_is_active ON mcp_tokens(is_active);
