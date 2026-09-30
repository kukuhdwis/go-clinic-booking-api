package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Clinic struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Address   string         `gorm:"type:text" json:"address"`
	Phone     string         `gorm:"type:varchar(20)" json:"phone"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Clinic) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type CreateClinicRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=100"`
	Address string `json:"address" binding:"required"`
	Phone   string `json:"phone" binding:"required"`
}

type ClinicRepository interface {
	Create(ctx context.Context, clinic *Clinic) error
	FindByID(ctx context.Context, id uuid.UUID) (*Clinic, error)
	FindAll(ctx context.Context) ([]Clinic, error)
}

type ClinicUsecase interface {
	CreateClinic(ctx context.Context, req *CreateClinicRequest) (*Clinic, error)
	GetClinics(ctx context.Context) ([]Clinic, error)
	GetClinicByID(ctx context.Context, id uuid.UUID) (*Clinic, error)
}
