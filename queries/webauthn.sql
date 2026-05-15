-- +goose StatementBegin

-- name: CreateWebAuthnCredential :one
INSERT INTO webauthn_credentials (user_id, credential_id, public_key, sign_count)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWebAuthnCredentialsByUserID :many
SELECT * FROM webauthn_credentials
WHERE user_id = $1;

-- name: GetWebAuthnCredentialByCredentialID :one
SELECT * FROM webauthn_credentials
WHERE credential_id = $1;

-- name: UpdateWebAuthnCredentialSignCount :exec
UPDATE webauthn_credentials
SET sign_count = $2
WHERE credential_id = $1;

-- name: CreateWebAuthnSession :one
INSERT INTO webauthn_sessions (data, expires_at)
VALUES ($1, $2)
RETURNING *;

-- name: GetWebAuthnSessionByID :one
SELECT * FROM webauthn_sessions
WHERE id = $1
  AND expires_at > NOW();

-- name: DeleteWebAuthnSession :exec
DELETE FROM webauthn_sessions WHERE id = $1;

-- name: DeleteExpiredWebAuthnSessions :exec
DELETE FROM webauthn_sessions WHERE expires_at <= NOW();

-- +goose StatementEnd
