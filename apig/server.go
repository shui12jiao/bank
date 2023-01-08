package apig

import (
	"fmt"

	db "github.com/shui12jiao/my_simplebank/db/sqlc"
	"github.com/shui12jiao/my_simplebank/pb"
	"github.com/shui12jiao/my_simplebank/tasks"
	"github.com/shui12jiao/my_simplebank/token"
	"github.com/shui12jiao/my_simplebank/util"
)

// Server servers gRPC requests for our banking service
type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor tasks.TaskDistributor
}

// NewServer creates a new HTTP server and setup routing
func NewServer(config util.Config, store db.Store, taskDistributor tasks.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}
