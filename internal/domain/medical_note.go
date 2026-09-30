package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MedicalNote struct {
	ID            uuid.UUID      `gorm:"type:char(36);primary_key" json:"id"`
	AppointmentID uuid.UUID      `gorm:"type:char(36);not null;uniqueIndex" json:"appointment_id"`
	Diagnosis     string         `gorm:"type:text;not null" json:"diagnosis"`
	Prescription  string         `gorm:"type:text" json:"prescription"`
	Notes         string         `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *MedicalNote) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
