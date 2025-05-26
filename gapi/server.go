package gapi

import (
	"fmt"

	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/pb"
	"github.com/starjardin/simplebank/token"
	"github.com/starjardin/simplebank/utils"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store      db.Store
	tokenMaker token.Maker
	config     utils.Config
}

func NewServer(config utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		store:      store,
		config:     config,
		tokenMaker: tokenMaker,
	}
	return server, nil
}
