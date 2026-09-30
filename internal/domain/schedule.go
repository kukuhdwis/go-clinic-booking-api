package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DoctorSchedule struct {
	ID        uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	DoctorID  uuid.UUID      `gorm:"type:char(36);not null;index" json:"doctor_id"`
	DayOfWeek int            `gorm:"type:tinyint;not null" json:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	StartTime string         `gorm:"type:varchar(8);not null" json:"start_time"` // Format "15:04" or "15:04:05"
	EndTime   string         `gorm:"type:varchar(8);not null" json:"end_time"`   // Format "15:04" or "15:04:05"
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *DoctorSchedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type CreateScheduleRequest struct {
	DoctorID  uuid.UUID `json:"doctor_id" binding:"required"`
	DayOfWeek int       `json:"day_of_week" binding:"min=0,max=6"`
	StartTime string    `json:"start_time" binding:"required"` // "09:00"
	EndTime   string    `json:"end_time" binding:"required"`   // "17:00"
	IsActive  *bool     `json:"is_active"`
}

type DoctorCreateScheduleRequest struct {
	DayOfWeek int    `json:"day_of_week" binding:"min=0,max=6"`
	StartTime string `json:"start_time" binding:"required"` // "09:00" or "09:00:00"
	EndTime   string `json:"end_time" binding:"required"`   // "17:00" or "17:00:00"
	IsActive  *bool  `json:"is_active"`
}

type AvailableSlot struct {
	StartTime   string `json:"start_time"` // "09:00:00"
	EndTime     string `json:"end_time"`   // "09:30:00"
	IsAvailable bool   `json:"is_available"`
}

type DoctorAvailabilityResponse struct {
	DoctorID       uuid.UUID       `json:"doctor_id"`
	DoctorName     string          `json:"doctor_name"`
	Specialization string          `json:"specialization"`
	Date           string          `json:"date"`
	DayOfWeek      int             `json:"day_of_week"`
	SlotDuration   int             `json:"slot_duration_minutes"`
	Slots          []AvailableSlot `json:"slots"`
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *DoctorSchedule) error
	FindByID(ctx context.Context, id uuid.UUID) (*DoctorSchedule, error)
	FindByDoctorAndDay(ctx context.Context, doctorID uuid.UUID, dayOfWeek int) ([]DoctorSchedule, error)
	FindByDoctorID(ctx context.Context, doctorID uuid.UUID) ([]DoctorSchedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ScheduleUsecase interface {
	CreateSchedule(ctx context.Context, req *CreateScheduleRequest) (*DoctorSchedule, error)
	CreateDoctorOwnSchedule(ctx context.Context, doctorUserID uuid.UUID, req *DoctorCreateScheduleRequest) (*DoctorSchedule, error)
	GetDoctorSchedules(ctx context.Context, doctorID uuid.UUID) ([]DoctorSchedule, error)
	GetDoctorOwnSchedules(ctx context.Context, doctorUserID uuid.UUID) ([]DoctorSchedule, error)
	DeleteDoctorSchedule(ctx context.Context, doctorUserID uuid.UUID, scheduleID uuid.UUID) error
	GetAvailableSlots(ctx context.Context, doctorID uuid.UUID, targetDate string) (*DoctorAvailabilityResponse, error)
}
