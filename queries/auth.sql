-- +goose StatementBegin

-- name: UpsertUser :one
INSERT INTO users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: CreateMagicLink :one
INSERT INTO magic_links (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMagicLinkByToken :one
SELECT magic_links.*, users.email
FROM magic_links
JOIN users ON users.id = magic_links.user_id
WHERE magic_links.token = $1
  AND magic_links.expires_at > NOW();

-- name: DeleteMagicLink :exec
DELETE FROM magic_links WHERE id = $1;

-- name: DeleteExpiredMagicLinks :exec
DELETE FROM magic_links WHERE expires_at <= NOW();

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, csrf_token, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSessionByID :one
SELECT sessions.*, users.email
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.id = $1
  AND sessions.expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= NOW();

-- +goose StatementEnd
