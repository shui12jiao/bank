package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	accountF := createRandomAccount(t)
	accountT := createRandomAccount(t)

	//run a concurrent transfer transacton
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: accountF.ID,
				ToAccountID:   accountT.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	//check results
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		//check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccountsID, accountF.ID)
		require.Equal(t, transfer.ToAccountsID, accountT.ID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//check entries
		entryF := result.FromEntry
		require.NotEmpty(t, entryF)
		require.Equal(t, entryF.AccountsID, accountF.ID)
		require.Equal(t, entryF.Amount, -amount)
		require.NotZero(t, entryF.ID)
		require.NotZero(t, entryF.CreatedAt)

		_, err = store.GetEntry(context.Background(), entryF.ID)
		require.NoError(t, err)

		entryT := result.ToEntry
		require.NotEmpty(t, entryT)
		require.Equal(t, entryT.AccountsID, accountT.ID)
		require.Equal(t, entryT.Amount, amount)
		require.NotZero(t, entryT.ID)
		require.NotZero(t, entryT.CreatedAt)

		_, err = store.GetEntry(context.Background(), entryT.ID)
		require.NoError(t, err)

		//check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, accountF.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, accountT.ID)

		//check accounts'balance
		diff1 := accountF.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - accountT.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	//check the final updated balances
	updatedAccountF, err := testQueries.GetAccount(context.Background(), accountF.ID)
	require.NoError(t, err)

	updatedAccountT, err := testQueries.GetAccount(context.Background(), accountT.ID)
	require.NoError(t, err)

	require.Equal(t, updatedAccountF.Balance, accountF.Balance-amount*int64(n))
	require.Equal(t, updatedAccountT.Balance, accountT.Balance+amount*int64(n))
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	//run a concurrent transfer transacton
	n := 10
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		ToAccountID := account2.ID
		if i%2 == 1 {
			fromAccountID = account2.ID
			ToAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   ToAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	//check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	//check the final updated balances
	updatedAccountF, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccountT, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, updatedAccountF.Balance, account1.Balance)
	require.Equal(t, updatedAccountT.Balance, account2.Balance)
}
