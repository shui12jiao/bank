package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/shui12jiao/my_simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreatUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUserAnyField(t *testing.T) {
	oldUser := createRandomUser(t)

	hashedPassword := util.RandomString(6)
	fullName := util.RandomOwner()
	email := util.RandomEmail()

	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
		Username:       oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, newUser)
	require.Equal(t, hashedPassword, newUser.HashedPassword)
	require.Equal(t, oldUser.Username, newUser.Username)
	require.Equal(t, oldUser.FullName, newUser.FullName)
	require.Equal(t, oldUser.Email, newUser.Email)
	oldUser = newUser

	newUser, err = testQueries.UpdateUser(context.Background(), UpdateUserParams{
		FullName: sql.NullString{String: fullName, Valid: true},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, newUser)
	require.Equal(t, fullName, newUser.FullName)
	require.Equal(t, oldUser.Username, newUser.Username)
	require.Equal(t, oldUser.HashedPassword, newUser.HashedPassword)
	require.Equal(t, oldUser.Email, newUser.Email)
	oldUser = newUser

	newUser, err = testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Email:    sql.NullString{String: email, Valid: true},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, newUser)
	require.Equal(t, email, newUser.Email)
	require.Equal(t, oldUser.Username, newUser.Username)
	require.Equal(t, oldUser.HashedPassword, newUser.HashedPassword)
	require.Equal(t, oldUser.FullName, newUser.FullName)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)
	arg := UpdateUserParams{
		Username:       oldUser.Username,
		HashedPassword: sql.NullString{String: util.RandomString(6), Valid: true},
		FullName:       sql.NullString{String: util.RandomOwner(), Valid: true},
		Email:          sql.NullString{String: util.RandomEmail(), Valid: true},
	}
	newUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, newUser)
	require.Equal(t, arg.Username, newUser.Username)
	require.Equal(t, arg.HashedPassword.String, newUser.HashedPassword)
	require.Equal(t, arg.FullName.String, newUser.FullName)
	require.Equal(t, arg.Email.String, newUser.Email)
}
