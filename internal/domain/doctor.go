package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Doctor struct {
	ID                  uuid.UUID        `gorm:"type:char(36);primary_key" json:"id"`
	UserID              uuid.UUID        `gorm:"type:char(36);not null;uniqueIndex" json:"user_id"`
	ClinicID            uuid.UUID        `gorm:"type:char(36);not null;index" json:"clinic_id"`
	Specialization      string           `gorm:"type:varchar(100);not null;index" json:"specialization"`
	SlotDurationMinutes int              `gorm:"type:int;default:30" json:"slot_duration_minutes"`
	ConsultationFee     float64          `gorm:"type:decimal(12,2);default:0.00" json:"consultation_fee"`
	Bio                 string           `gorm:"type:text" json:"bio"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	DeletedAt           gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	User      User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Clinic    Clinic           `gorm:"foreignKey:ClinicID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"clinic,omitempty"`
	Schedules []DoctorSchedule `gorm:"foreignKey:DoctorID" json:"schedules,omitempty"`
}

func (d *Doctor) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.SlotDurationMinutes <= 0 {
		d.SlotDurationMinutes = 30
	}
	return nil
}

type CreateDoctorRequest struct {
	// Option A: Link existing user by ID
	UserID uuid.UUID `json:"user_id"`

	// Option B: Atomic unified creation (creates user account with doctor role)
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
	FullName string `json:"full_name" binding:"omitempty,min=2,max=100"`
	Phone    string `json:"phone" binding:"omitempty,e164|numeric"`

	// Professional & Clinic Assignment
	ClinicID            uuid.UUID `json:"clinic_id" binding:"required"`
	Specialization      string    `json:"specialization" binding:"required,min=2,max=100"`
	SlotDurationMinutes int       `json:"slot_duration_minutes" binding:"omitempty,min=10,max=120"`
	ConsultationFee     float64   `json:"consultation_fee" binding:"gte=0"`
	Bio                 string    `json:"bio"`
}

type UpdateDoctorProfileRequest struct {
	Specialization      *string  `json:"specialization" binding:"omitempty,min=2,max=100"`
	SlotDurationMinutes *int     `json:"slot_duration_minutes" binding:"omitempty,min=10,max=120"`
	ConsultationFee     *float64 `json:"consultation_fee" binding:"omitempty,gte=0"`
	Bio                 *string  `json:"bio"`
}

type DoctorFilter struct {
	Specialization string    `form:"specialization"`
	ClinicID       uuid.UUID `form:"clinic_id"`
	Search         string    `form:"search"` // searches doctor name or specialization
	Page           int       `form:"page,default=1"`
	Limit          int       `form:"limit,default=10"`
}

type DoctorRepository interface {
	Create(ctx context.Context, doctor *Doctor) error
	Update(ctx context.Context, doctor *Doctor) error
	FindByID(ctx context.Context, id uuid.UUID) (*Doctor, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*Doctor, error)
	FindAll(ctx context.Context, filter DoctorFilter) ([]Doctor, int64, error)
}

type DoctorUsecase interface {
	CreateDoctor(ctx context.Context, req *CreateDoctorRequest) (*Doctor, error)
	GetDoctorByID(ctx context.Context, id uuid.UUID) (*Doctor, error)
	GetDoctorByUserID(ctx context.Context, userID uuid.UUID) (*Doctor, error)
	UpdateDoctorProfile(ctx context.Context, userID uuid.UUID, req *UpdateDoctorProfileRequest) (*Doctor, error)
	GetDoctors(ctx context.Context, filter DoctorFilter) ([]Doctor, int64, error)
}
