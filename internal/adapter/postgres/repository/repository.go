package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/synthao/sso/internal/domain"
)

var (
	ErrGetUser     = errors.New("failed to get user")
	ErrCreateToken = errors.New("failed to create token")
	ErrUpdateToken = errors.New("failed to update token")
)

type Repository interface {
	GetUser(nickname string) (*domain.User, error)
	GetTokenByRefreshToken(refreshToken string) (*domain.Token, error)
	CreateToken(refreshToken string, userID int) error
	UpdateToken(refreshToken string, userID int) error
}

type user struct {
	ID       int    `db:"id"`
	Status   int    `db:"status"`
	Nickname string `db:"nickname"`
	Password string `db:"password"`
}

type token struct {
	ID     int `db:"id"`
	UserID int `db:"user_id"`
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetUser(nickname string) (*domain.User, error) {
	var dest user

	err := r.db.Get(&dest, "SELECT id, status, nickname, password FROM users WHERE nickname = $1", nickname)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w, %w", ErrGetUser, err)
		}

		return nil, err
	}

	return &domain.User{
		ID:       dest.ID,
		Nickname: dest.Nickname,
		Password: dest.Password,
	}, nil
}

func (r *repository) GetTokenByRefreshToken(refreshToken string) (*domain.Token, error) {
	var dest token

	err := r.db.Get(&dest, "SELECT id, user_id FROM user_tokens WHERE refresh_token = $1", refreshToken)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w, %w", ErrGetUser, err)
		}

		return nil, err
	}

	return &domain.Token{ID: dest.ID, UserID: dest.UserID}, nil
}

func (r *repository) CreateToken(refreshToken string, userID int) error {
	_, err := r.db.Exec("INSERT INTO user_tokens (user_id, refresh_token) VALUES ($1, $2)", userID, refreshToken)
	if err != nil {
		return fmt.Errorf("%w, %w", ErrCreateToken, err)
	}

	return nil
}

func (r *repository) UpdateToken(refreshToken string, userID int) error {
	_, err := r.db.Exec("UPDATE user_tokens SET refresh_token = $1, updated_at = now() WHERE user_id = $2", refreshToken, userID)
	if err != nil {
		return fmt.Errorf("%w, %w", ErrUpdateToken, err)
	}

	return nil
}
