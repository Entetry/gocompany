package middleware

import (
	"context"
	"fmt"
	"github.com/Entetry/gocompany/internal/config"
	"github.com/Entetry/gocompany/internal/model"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	cfg *config.JwtConfig
}

// NewAuthInterceptor creates jwt middleware object
func NewAuthInterceptor(cfg *config.JwtConfig) *AuthInterceptor {
	return &AuthInterceptor{cfg: cfg}
}

func (a *AuthInterceptor) Unary(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Println(info.FullMethod)
	if info.FullMethod == "/proto.AuthGRPCService/SingUp" || info.FullMethod == "/proto.AuthGRPCService/SignIn" {
		return handler(ctx, req)
	}
	err := a.authorize(ctx)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (a *AuthInterceptor) StreamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	log.Println("--> stream auth interceptor: ", info.FullMethod)
	err := a.authorize(stream.Context())
	if err != nil {
		return err
	}
	return handler(srv, stream)
}

func (a *AuthInterceptor) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
	}
	authHeader, ok := md["authorization"]
	if !ok {
		return status.Errorf(codes.Unauthenticated, "Authorization token is not supplied")
	}
	token := authHeader[0]
	_, err := a.validateToken(token)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, err.Error())
	}
	return nil
}

func (a *AuthInterceptor) validateToken(accessToken string) (*model.Claim, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&model.Claim{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(a.cfg.AccessTokenKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*model.Claim)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
