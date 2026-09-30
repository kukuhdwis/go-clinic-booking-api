package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"projectgo/internal/domain"
	"projectgo/pkg/password"
	"projectgo/pkg/token"
)

var (
	ErrEmailAlreadyExists = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

type authUsecase struct {
	userRepo   domain.UserRepository
	jwtManager *token.JWTManager
}

func NewAuthUsecase(userRepo domain.UserRepository, jwtManager *token.JWTManager) domain.AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (u *authUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Check if email already in use
	existing, _ := u.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Role:         domain.RolePatient, // Default registration role is always patient
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	tokenStr, expiresAt, err := u.jwtManager.GenerateToken(newUser.ID, newUser.Email, string(newUser.Role))
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
		User: domain.UserBrief{
			ID:       newUser.ID,
			Email:    newUser.Email,
			FullName: newUser.FullName,
			Role:     newUser.Role,
		},
	}, nil
}

func (u *authUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !password.Check(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	tokenStr, expiresAt, err := u.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
		User: domain.UserBrief{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
		},
	}, nil
}

func (u *authUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
