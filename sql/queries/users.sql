-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetLastUser :one
SELECT *
FROM users
ORDER BY created_at ASC
LIMIT 1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: UpdateUserByID :one
UPDATE users SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpgradeToChirpyRed :one
UPDATE users
SET
  is_chirpy_red = TRUE,
  updated_at     = NOW()
WHERE id = $1
RETURNING *;


-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1;