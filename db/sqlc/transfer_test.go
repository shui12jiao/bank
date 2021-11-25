package db

import (
	"context"
	"testing"
	"time"

	"github.com/shui12jiao/my_simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, accountF, accountT Account) Transfer {
	arg := CreateTransferParams{
		FromAccountsID: accountF.ID,
		ToAccountsID:   accountT.ID,
		Amount:         util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountsID, transfer.FromAccountsID)
	require.Equal(t, arg.ToAccountsID, transfer.ToAccountsID)
	require.Equal(t, arg.Amount, transfer.Amount)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreatTransfer(t *testing.T) {
	accountF := createRandomAccount(t)
	accountT := createRandomAccount(t)
	createRandomTransfer(t, accountF, accountT)
}

func TestGetTransfer(t *testing.T) {
	accountF := createRandomAccount(t)
	accountT := createRandomAccount(t)
	transfer1 := createRandomTransfer(t, accountF, accountT)
	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountsID, transfer2.FromAccountsID)
	require.Equal(t, transfer1.ToAccountsID, transfer2.ToAccountsID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}

func TestListTransfer(t *testing.T) {
	accountF := createRandomAccount(t)
	accountT := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomTransfer(t, accountF, accountT)
	}

	argF := ListTransferParams{
		FromAccountsID: accountF.ID,
		Limit:          5,
		Offset:         5,
	}

	argT := ListTransferParams{
		ToAccountsID: accountT.ID,
		Limit:        5,
		Offset:       5,
	}

	transfersF, err := testQueries.ListTransfer(context.Background(), argF)
	require.NoError(t, err)
	require.Len(t, transfersF, 5)

	for _, transfer := range transfersF {
		require.NotEmpty(t, transfer)
	}

	transfersT, err := testQueries.ListTransfer(context.Background(), argT)
	require.NoError(t, err)
	require.Len(t, transfersT, 5)

	for _, transfer := range transfersT {
		require.NotEmpty(t, transfer)
	}

}
