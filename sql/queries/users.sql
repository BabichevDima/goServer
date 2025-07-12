-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES (
    $1,
    $2
    )
RETURNING id, created_at, updated_at, email;

-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
VALUES (
    $1,
    $2
    )
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;