-- Drop indexes
DROP INDEX IF EXISTS idx_mcp_tokens_is_active;
DROP INDEX IF EXISTS idx_mcp_tokens_user_id;
DROP INDEX IF EXISTS idx_mcp_tokens_token;

-- Drop mcp_tokens table
DROP TABLE IF EXISTS mcp_tokens;
