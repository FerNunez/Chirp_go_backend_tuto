-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password, is_chirpy_red)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2,
  FALSE
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserLoginByID :exec
UPDATE users
SET email = $1,
    hashed_password = $2,
    updated_at = NOW()
WHERE id = $3;

-- name: ResetUsers :exec
TRUNCATE TABLE users RESTART IDENTITY CASCADE;

-- name: UpdateChirpyRedByID :exec
UPDATE users
SET is_chirpy_red = $1
WHERE id = $2;

