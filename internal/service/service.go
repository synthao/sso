package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/synthao/sso/gen/go/sso"
	"github.com/synthao/sso/internal/adapter/postgres/repository"
	"github.com/synthao/sso/internal/auth"
	"github.com/synthao/sso/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SSOService struct {
	sso_v1.UnimplementedServiceServer
	repository repository.Repository
	logger     *zap.Logger
	config     *config.JWT
}

func NewSSOService(repository repository.Repository, logger *zap.Logger, config *config.JWT) *SSOService {
	return &SSOService{repository: repository, logger: logger, config: config}
}

func (s *SSOService) Authenticate(ctx context.Context, req *sso_v1.AuthenticateRequest) (*sso_v1.AuthenticateResponse, error) {
	user, err := s.repository.GetUser(req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
		}

		s.logger.Error("failed to get user", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "oops, something went wrong")
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	ok, err := auth.Verify(req.Password, user.Password)
	if err != nil {
		s.logger.Error("failed to verify password", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	accessToken, refreshToken, err := auth.GenerateTokens(s.config.Secret, user.ID)
	if err != nil {
		s.logger.Error("failed to get user", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	err = s.repository.CreateToken(refreshToken, user.ID)
	if err != nil {
		s.logger.Error("failed to create token", zap.Error(err), zap.Int("user_id", user.ID))

		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	return &sso_v1.AuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *SSOService) Authorize(ctx context.Context, req *sso_v1.AuthorizeRequest) (*sso_v1.AuthorizeResponse, error) {
	claims, err := auth.ValidateJWTToken(s.config.Secret, req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	if claims == nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect token: %v", err)
	}

	return &sso_v1.AuthorizeResponse{
		UserId: int32(claims.UserID),
	}, nil
}

func (s *SSOService) Refresh(ctx context.Context, req *sso_v1.RefreshTokenRequest) (*sso_v1.RefreshTokenResponse, error) {
	token, err := s.repository.GetTokenByRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.Unauthenticated, "no refresh token")
		}

		s.logger.Error("failed to get token by refresh token", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "oops, something went wrong")
	}
	if token == nil {
		return nil, status.Errorf(codes.Unauthenticated, "no refresh token")
	}

	accessToken, refreshToken, err := auth.GenerateTokens(s.config.Secret, token.UserID)
	if err != nil {
		s.logger.Error("failed generate tokens when refresh token", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	err = s.repository.UpdateToken(refreshToken, token.UserID)
	if err != nil {
		s.logger.Error("failed update token when refresh token", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "failed to update token: %v", err)
	}

	return &sso_v1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *SSOService) IsAuthorized(ctx context.Context, req *sso_v1.IsAuthorizedRequest) (*sso_v1.IsAuthorizedResponse, error) {
	claims, err := auth.ValidateJWTToken(s.config.Secret, req.Token)
	if err != nil {
		s.logger.Error("failed to validate jwt token", zap.Error(err))
		return &sso_v1.IsAuthorizedResponse{IsAuthorized: false}, nil
	}
	if claims == nil {
		return &sso_v1.IsAuthorizedResponse{IsAuthorized: false}, nil
	}

	return &sso_v1.IsAuthorizedResponse{IsAuthorized: true}, nil
}
