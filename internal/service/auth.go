package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"entetry/gotest/internal/config"
	"entetry/gotest/internal/model"
)

const (
	wrongPassword         = "wrong password"
	refreshTokenIsExpired = "refresh token is expired"
	invalidFingerprint    = "invalid fingerprint"
	userAlreadyExist      = "user already exists"
)

// AuthService service interface
type AuthService interface {
	SignIn(ctx context.Context, username, password string, tokenParam *model.TokenParam) (refreshToken, accessToken string, err error)
	SignUp(ctx context.Context, username, password, email string) error
	RefreshTokens(ctx context.Context, refreshToken string, tokenParam *model.TokenParam) (newRefreshToken, accessToken string, err error)
	Logout(ctx context.Context, refreshToken string) (err error)
}

// Auth service struct
type Auth struct {
	userService    *User
	refreshSession *RefreshSession
	cfg            *config.JwtConfig
}

// NewAuthService creates new Auth service
func NewAuthService(userService *User, refreshSession *RefreshSession, cfg *config.JwtConfig) *Auth {
	return &Auth{
		userService:    userService,
		refreshSession: refreshSession,
		cfg:            cfg}
}

// SignIn sign in user and return tokens
func (a *Auth) SignIn(ctx context.Context, username, password string, tokenParam *model.TokenParam) (refreshToken, accessToken string, err error) {
	user, err := a.attemptLogin(ctx, username, password)
	if err != nil {
		return "", "", fmt.Errorf(wrongPassword)
	}
	return a.generateTokens(ctx, user.ID.String(), tokenParam)
}

// SignUp  sign up new user
func (a *Auth) SignUp(ctx context.Context, username, password, email string) error {
	user, err := a.userService.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	if user != nil {
		return fmt.Errorf(userAlreadyExist)
	}

	_, err = a.userService.Create(ctx, username, password, email)
	if err != nil {
		return err
	}
	return nil
}

// RefreshTokens update refresh token
func (a *Auth) RefreshTokens(ctx context.Context, refreshToken string,
	tokenParam *model.TokenParam) (newRefreshToken, accessToken string, err error) {
	session, err := a.refreshSession.PopSession(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	if session.ExpiresAt <= time.Now().Unix() {
		err = a.refreshSession.DeleteUserSessions(ctx, session.UserID)
		if err != nil {
			log.Error(err)
		}
		return "", "", fmt.Errorf(refreshTokenIsExpired)
	}

	if !a.checkFingerprint(session, tokenParam) {
		err = a.refreshSession.DeleteUserSessions(ctx, session.UserID)
		if err != nil {
			log.Error(err)
		}
		return "", "", fmt.Errorf(invalidFingerprint)
	}

	return a.generateTokens(ctx, session.UserID, tokenParam)
}

// Logout log out from session
func (a *Auth) Logout(ctx context.Context, refreshToken string) (err error) {
	err = a.refreshSession.Delete(ctx, refreshToken)
	if err != nil {
		return err
	}
	return nil
}

func (a *Auth) checkFingerprint(session *model.RefreshSession, tokenParam *model.TokenParam) bool {
	return tokenParam.IP == session.TokenParam.IP ||
		tokenParam.UserAgent == session.UserAgent ||
		tokenParam.Fingerprint == session.Fingerprint
}

func (a *Auth) attemptLogin(ctx context.Context, username, password string) (*model.User, error) {
	user, err := a.userService.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *Auth) generateTokens(ctx context.Context, userID string, tokenParam *model.TokenParam) (refreshToken, accessToken string, err error) {
	refreshToken = uuid.New().String()
	err = a.refreshSession.SaveSession(ctx, &model.RefreshSession{
		RefreshToken: refreshToken,
		UserID:       userID,
		ExpiresAt:    time.Now().Add(a.cfg.RefreshTokenExpiration).Unix(),
		TokenParam:   *tokenParam,
	})
	if err != nil {
		return "", "", err
	}
	accessToken, err = a.generateAccessToken(userID, a.cfg.AccessTokenKey, time.Now().Add(a.cfg.AccessTokenExpiration).Unix())
	if err != nil {
		return "", "", err
	}

	return refreshToken, accessToken, nil
}

func (a *Auth) generateAccessToken(userID, key string, expiresAt int64) (string, error) {
	claim := model.Claim{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expiresAt,
		},
		UserID: userID,
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claim).SignedString([]byte(key))
	if err != nil {
		return "", fmt.Errorf("error in SignedString for userID: %v and key: %v", userID, key)
	}

	return token, err
}
