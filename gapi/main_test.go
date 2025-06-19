package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/token"
	"github.com/starjardin/simplebank/utils"
	"github.com/starjardin/simplebank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := utils.Config{
		TokenSymetricKey:    utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor)

	require.NoError(t, err)

	return server
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, role string, duration time.Duration) context.Context {
	ctx := context.Background()

	accessToken, _, err := tokenMaker.CreateToken(username, role, duration)

	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)

	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(ctx, md)
}
