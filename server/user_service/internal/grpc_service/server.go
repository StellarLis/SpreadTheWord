package grpc_service

import (
	"context"
	authv1 "user_service/pkg/grpc/auth"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
)

type serverAPI struct {
	authv1.UnimplementedAuthServer
	JwtSecretKey string
}

func Register(gRPC *grpc.Server, jwtSecretKey string) {
	authv1.RegisterAuthServer(gRPC, &serverAPI{JwtSecretKey: jwtSecretKey})
}

func (s *serverAPI) Authenticate(
	ctx context.Context,
	req *authv1.AuthenticateRequest,
) (*authv1.AuthenticateResponse, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(req.GetToken(), claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.JwtSecretKey), nil
	})
	if err != nil {
		return &authv1.AuthenticateResponse{
			StatusCode: 401,
		}, nil
	}

	return &authv1.AuthenticateResponse{
		StatusCode: 200,
		UserId:     int64(claims["userId"].(float64)),
		Username:   claims["username"].(string),
	}, nil
}
