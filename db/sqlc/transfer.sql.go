// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: transfer.sql

package db

import (
	"context"
)

const createTransfer = `-- name: CreateTransfer :one
INSERT INTO transfers (
  from_accounts_id,
  to_accounts_id,
  amount
) VALUES (
  $1, $2, $3
) RETURNING id, from_accounts_id, to_accounts_id, amount, created_at
`

type CreateTransferParams struct {
	FromAccountsID int64 `json:"fromAccountsID"`
	ToAccountsID   int64 `json:"toAccountsID"`
	Amount         int64 `json:"amount"`
}

func (q *Queries) CreateTransfer(ctx context.Context, arg CreateTransferParams) (Transfer, error) {
	row := q.db.QueryRowContext(ctx, createTransfer, arg.FromAccountsID, arg.ToAccountsID, arg.Amount)
	var i Transfer
	err := row.Scan(
		&i.ID,
		&i.FromAccountsID,
		&i.ToAccountsID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const getTransfer = `-- name: GetTransfer :one
SELECT id, from_accounts_id, to_accounts_id, amount, created_at FROM transfers
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetTransfer(ctx context.Context, id int64) (Transfer, error) {
	row := q.db.QueryRowContext(ctx, getTransfer, id)
	var i Transfer
	err := row.Scan(
		&i.ID,
		&i.FromAccountsID,
		&i.ToAccountsID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const listTransfer = `-- name: ListTransfer :many
SELECT id, from_accounts_id, to_accounts_id, amount, created_at FROM transfers
WHERE
    from_accounts_id = $1 OR
    to_accounts_id = $2
ORDER BY id
LIMIT $3
OFFSET $4
`

type ListTransferParams struct {
	FromAccountsID int64 `json:"fromAccountsID"`
	ToAccountsID   int64 `json:"toAccountsID"`
	Limit          int32 `json:"limit"`
	Offset         int32 `json:"offset"`
}

func (q *Queries) ListTransfer(ctx context.Context, arg ListTransferParams) ([]Transfer, error) {
	rows, err := q.db.QueryContext(ctx, listTransfer,
		arg.FromAccountsID,
		arg.ToAccountsID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Transfer{}
	for rows.Next() {
		var i Transfer
		if err := rows.Scan(
			&i.ID,
			&i.FromAccountsID,
			&i.ToAccountsID,
			&i.Amount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
