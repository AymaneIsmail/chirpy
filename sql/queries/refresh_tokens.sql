-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  token, created_at, updated_at, user_id, expires_at, revoked_at
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshTokenByToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: GetUserByRefreshToken :one
SELECT users.*
FROM refresh_tokens
JOIN users ON refresh_tokens.user_id = users.id
WHERE refresh_tokens.token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1;