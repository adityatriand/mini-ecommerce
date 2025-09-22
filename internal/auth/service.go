package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo *Repository
	rdb  *redis.Client
}

func NewService(repo *Repository, rdb *redis.Client) *Service {
	return &Service{repo: repo, rdb: rdb}
}

func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	hashed, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	user := User{
		Email:    email,
		Password: hashed,
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return User{}, "", err
	}

	if !CheckPassword(user.Password, password) {
		return User{}, "", ErrInvalidCredentials
	}

	sessionID := uuid.New().String()
	if err := s.rdb.Set(ctx, "session:"+sessionID, user.ID, 3600*time.Second).Err(); err != nil {
		return User{}, "", err
	}

	return user, sessionID, nil
}

func (s *Service) RegisterUser(ctx context.Context, input RegisterInput) (*User, error) {
	hashed, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := User{
		Email:    input.Email,
		Password: hashed,
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) LoginUser(ctx context.Context, input LoginInput) (*User, string, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, "", err
	}

	if !CheckPassword(user.Password, input.Password) {
		return nil, "", ErrInvalidCredentials
	}

	sessionID := uuid.New().String()
	if err := s.rdb.Set(ctx, "session:"+sessionID, user.ID, 3600*time.Second).Err(); err != nil {
		return nil, "", err
	}

	return &user, sessionID, nil
}
