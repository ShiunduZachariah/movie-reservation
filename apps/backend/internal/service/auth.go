package service

import (
	"context"
	"time"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/errs"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/model"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type AuthService struct {
	db     *pgxpool.Pool
	users  *repository.UserRepository
	secret string
	logger *zerolog.Logger
}

type AuthResult struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func NewAuthService(db *pgxpool.Pool, users *repository.UserRepository, secret string, logger *zerolog.Logger) *AuthService {
	return &AuthService{db: db, users: users, secret: secret, logger: logger}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*AuthResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	user, err := s.users.CreateLocal(ctx, tx, email, name, password, model.RoleUser)
	if err != nil {
		s.logger.Warn().Err(err).Str("email", email).Msg("user registration failed")
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("user registration commit failed")
		return nil, err
	}

	token, err := s.issueToken(user)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("user registration token issue failed")
		return nil, err
	}

	s.logger.Info().
		Str("event", "user_registered").
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Str("role", string(user.Role)).
		Msg("user registered successfully")

	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.AuthenticateLocal(ctx, email, password)
	if err != nil {
		s.logger.Warn().Str("event", "login_failed").Str("email", email).Msg("user login failed")
		return nil, errs.Unauthorized("INVALID_CREDENTIALS", "invalid email or password")
	}

	token, err := s.issueToken(user)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("user login token issue failed")
		return nil, err
	}

	s.logger.Info().
		Str("event", "user_logged_in").
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Str("role", string(user.Role)).
		Msg("user login successful")

	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) issueToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ClerkID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", errs.Internal("failed to sign token")
	}
	return signed, nil
}
