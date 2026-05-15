-- +goose StatementBegin

-- name: CreateForm :one
INSERT INTO forms (user_id, title, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFormByID :one
SELECT * FROM forms WHERE id = $1;

-- name: ListFormsByUserID :many
SELECT * FROM forms
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: UpdateForm :one
UPDATE forms
SET title = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteForm :exec
DELETE FROM forms WHERE id = $1;

-- name: CreateQuestion :one
INSERT INTO questions (form_id, type, title, required, position, options, scale_min, scale_max)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListQuestionsByFormID :many
SELECT * FROM questions
WHERE form_id = $1
ORDER BY position;

-- name: UpdateQuestion :one
UPDATE questions
SET type = $2, title = $3, required = $4, position = $5, options = $6, scale_min = $7, scale_max = $8
WHERE id = $1
RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;

-- name: CreateResponse :one
INSERT INTO responses (form_id)
VALUES ($1)
RETURNING *;

-- name: ListResponsesByFormID :many
SELECT * FROM responses
WHERE form_id = $1
ORDER BY submitted_at DESC;

-- name: CountResponsesByFormID :one
SELECT COUNT(*) FROM responses WHERE form_id = $1;

-- name: CreateAnswer :one
INSERT INTO answers (response_id, question_id, value)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListAnswersByResponseID :many
SELECT * FROM answers WHERE response_id = $1;

-- +goose StatementEnd
