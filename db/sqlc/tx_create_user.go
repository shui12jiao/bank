package db

import "context"

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreateUser func(context.Context, User) error
}

type CreateUserTxResult struct {
	User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (result CreateUserTxResult, err error) {
	err = store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}
		return arg.AfterCreateUser(ctx, result.User)
	})

	return
}
