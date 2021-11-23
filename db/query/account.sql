-- name: CreateAccount :one
INSERT INTO accounts (
  owner,
  balance,
  curency
) VALUES (
  $1, $2, $3
) RETURNING *;
