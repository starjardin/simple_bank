package gapi

import (
	"fmt"

	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/pb"
	"github.com/starjardin/simplebank/token"
	"github.com/starjardin/simplebank/utils"
	"github.com/starjardin/simplebank/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store           db.Store
	tokenMaker      token.Maker
	config          utils.Config
	taskDistributor worker.TaskDistributor
}

func NewServer(config utils.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		store:           store,
		config:          config,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}
	return server, nil
}
