package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RolePatient Role = "patient"
	RoleDoctor  Role = "doctor"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Email        string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	FullName     string         `gorm:"type:varchar(100);not null" json:"full_name"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Role         Role           `gorm:"type:varchar(20);default:'patient';not null" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Request and Response DTOs
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
	Phone    string `json:"phone" binding:"omitempty,e164|numeric"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserBrief `json:"user"`
}

type UserBrief struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Role     Role      `json:"role"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*User, error)
}
