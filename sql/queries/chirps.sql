-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpsByID :one
SELECT * FROM chirps
WHERE id = ANY($1::UUID[]);

-- name: GetChirpsByUserID :many
SELECT * FROM chirps
WHERE user_id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1;
