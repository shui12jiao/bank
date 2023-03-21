package apig

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shui12jiao/my_simplebank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader     = "authorization"
	authorizationTypeBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata found in context")
	}
	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, errors.New("no authorization token found")
	}
	files := strings.Fields(values[0])
	if len(files) < 2 {
		return nil, errors.New("invalid authorization token format")
	}

	authorizationType := strings.ToLower(files[0])
	if authorizationType != authorizationTypeBearer {
		return nil, fmt.Errorf("unsupported authorization type %s", authorizationType)
	}

	payload, err := server.tokenMaker.VerifyToken(files[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return payload, nil
}
