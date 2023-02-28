package handlers

import (
	"context"
	"fmt"
	"github.com/Entetry/gocompany/protocol/companyService"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Entetry/gocompany/internal/model"
	"github.com/Entetry/gocompany/internal/service"
)

// Auth handler struct
type Auth struct {
	companyService.UnsafeAuthGRPCServiceServer
	authService *service.Auth
}

// NewAuth creates new auth handler
func NewAuth(authService *service.Auth) *Auth {
	return &Auth{authService: authService}
}

// SignIn sign in into account
func (a *Auth) SignIn(ctx context.Context, request *companyService.SignInRequest) (*companyService.SignInResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "Metadata error")
	}

	tokenParam, err := parseTokenParam(md)
	if err != nil {
		log.Error(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	refreshToken, accessToken, err := a.authService.SignIn(ctx, request.Username, request.Password, tokenParam)
	if err != nil {
		log.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &companyService.SignInResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// SignUp sign up account
func (a *Auth) SingUp(ctx context.Context, request *companyService.SignUpRequest) (*companyService.SignUpResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}

	if err := a.authService.SignUp(ctx, request.Username, request.Password, request.Email); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &companyService.SignUpResponse{}, nil
}

// Refresh update refresh token
func (a *Auth) Refresh(ctx context.Context, request *companyService.RefreshRequest) (*companyService.RefreshResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "Metadata error")
	}

	tokenParam, err := parseTokenParam(md)
	if err != nil {
		log.Error(err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	refreshToken, accessToken, err := a.authService.RefreshTokens(ctx, request.RefreshToken, tokenParam)
	if err != nil {
		log.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &companyService.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout log out from session
func (a *Auth) Logout(ctx context.Context, request *companyService.LogoutRequest) (*companyService.LogoutResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}

	if err := a.authService.Logout(ctx, request.RefreshToken); err != nil {
		log.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &companyService.LogoutResponse{}, nil
}

func parseTokenParam(metadata metadata.MD) (*model.TokenParam, error) {
	ua := metadata.Get("User-Agent")
	if len(ua) == 0 {
		return nil, fmt.Errorf("no User-Agent provided")
	}
	fingerprint := metadata.Get("Fingerprint")
	if len(fingerprint) == 0 {
		return nil, fmt.Errorf("no Fingerprint provided")
	}
	ip := metadata.Get("IP")
	if len(fingerprint) == 0 {
		return nil, fmt.Errorf("no IP provided")
	}

	return &model.TokenParam{
		UserAgent:   ua[0],
		Fingerprint: fingerprint[0],
		IP:          ip[0],
	}, nil
}
