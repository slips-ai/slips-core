-- name: UpsertUser :one
INSERT INTO users (user_id, username, avatar_url, email, updated_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
ON CONFLICT (user_id) DO UPDATE
SET 
    username = COALESCE(users.username, EXCLUDED.username),
    avatar_url = COALESCE(users.avatar_url, EXCLUDED.avatar_url),
    email = COALESCE(users.email, EXCLUDED.email),
    updated_at = CURRENT_TIMESTAMP
RETURNING id, user_id, username, avatar_url, email, created_at, updated_at;

-- name: GetUserByUserID :one
SELECT id, user_id, username, avatar_url, email, created_at, updated_at
FROM users
WHERE user_id = $1;

-- name: GetUserByID :one
SELECT id, user_id, username, avatar_url, email, created_at, updated_at
FROM users
WHERE id = $1;
