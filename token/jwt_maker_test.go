package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/starjardin/simplebank/utils"
	"github.com/stretchr/testify/require"
)

func JWTMakerTest(t *testing.T) {
	randomSecret := utils.RandomString(32)
	role := utils.DepositorRole
	maker, err := NewJWTMaker(randomSecret)
	require.NoError(t, err)

	username := utils.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(username, role, duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.Equal(t, role, payload.Role)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	randomSecret := utils.RandomString(32)
	role := utils.DepositorRole

	maker, err := NewJWTMaker(randomSecret)
	require.NoError(t, err)

	username := utils.RandomOwner()
	duration := time.Minute

	token, payload, err := maker.CreateToken(username, role, -duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)

	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := NewPayload(utils.RandomOwner(), utils.DepositorRole, time.Minute)

	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)

	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)

	require.NoError(t, err)

	maker, err := NewJWTMaker(utils.RandomString(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
