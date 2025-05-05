package token

import (
	"testing"
	"time"

	"github.com/starjardin/simplebank/utils"
	"github.com/stretchr/testify/require"
)

func PasetoMakerTest(t *testing.T) {
	randomSecret := utils.RandomString(32)

	maker, err := NewPasetoMaker(randomSecret)
	require.NoError(t, err)

	username := utils.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {
	randomSecret := utils.RandomString(32)

	maker, err := NewPasetoMaker(randomSecret)
	require.NoError(t, err)

	username := utils.RandomOwner()
	duration := time.Minute

	token, err := maker.CreateToken(username, -duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)

	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}
