package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCompleted AppointmentStatus = "completed"
	StatusCancelled AppointmentStatus = "cancelled"
)

type Appointment struct {
	ID                   uuid.UUID         `gorm:"type:char(36);primary_key" json:"id"`
	PatientID            uuid.UUID         `gorm:"type:char(36);not null;index" json:"patient_id"`
	DoctorID             uuid.UUID         `gorm:"type:char(36);not null;index:idx_doctor_slot,priority:1" json:"doctor_id"`
	AppointmentDate      string            `gorm:"type:date;not null;index:idx_doctor_slot,priority:2" json:"appointment_date"`
	StartTime            string            `gorm:"type:varchar(8);not null;index:idx_doctor_slot,priority:3" json:"start_time"`
	EndTime              string            `gorm:"type:varchar(8);not null" json:"end_time"`
	Status               AppointmentStatus `gorm:"type:varchar(20);default:'pending';not null;index" json:"status"`
	ComplaintDescription string            `gorm:"type:text" json:"complaint_description"`
	CancelReason         string            `gorm:"type:text" json:"cancel_reason,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	DeletedAt            gorm.DeletedAt    `gorm:"index" json:"-"`

	// Relationships
	Patient     User         `gorm:"foreignKey:PatientID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"patient,omitempty"`
	Doctor      Doctor       `gorm:"foreignKey:DoctorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"doctor,omitempty"`
	MedicalNote *MedicalNote `gorm:"foreignKey:AppointmentID" json:"medical_note,omitempty"`
}

func (a *Appointment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Status == "" {
		a.Status = StatusPending
	}
	return nil
}

type BookAppointmentRequest struct {
	DoctorID             uuid.UUID `json:"doctor_id" binding:"required"`
	AppointmentDate      string    `json:"appointment_date" binding:"required,datetime=2006-01-02"`
	StartTime            string    `json:"start_time" binding:"required"`
	ComplaintDescription string    `json:"complaint_description" binding:"required,min=5"`
}

type CancelAppointmentRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
}

type CompleteAppointmentRequest struct {
	Diagnosis    string `json:"diagnosis" binding:"required,min=3"`
	Prescription string `json:"prescription"`
	Notes        string `json:"notes"`
}

type AppointmentFilter struct {
	Status    AppointmentStatus `form:"status"`
	Date      string            `form:"date"`
	StartDate string            `form:"start_date"`
	EndDate   string            `form:"end_date"`
	DoctorID  uuid.UUID         `form:"doctor_id"`
	PatientID uuid.UUID         `form:"patient_id"`
	Page      int               `form:"page,default=1"`
	Limit     int               `form:"limit,default=10"`
}

type AppointmentRepository interface {
	CreateWithLock(ctx context.Context, appointment *Appointment) error
	FindByID(ctx context.Context, id uuid.UUID) (*Appointment, error)
	FindAll(ctx context.Context, filter AppointmentFilter) ([]Appointment, int64, error)
	FindActiveByDoctorAndDate(ctx context.Context, doctorID uuid.UUID, date string) ([]Appointment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status AppointmentStatus, cancelReason string) error
	CompleteWithMedicalNote(ctx context.Context, appointmentID uuid.UUID, note *MedicalNote) error
}

type AppointmentUsecase interface {
	BookAppointment(ctx context.Context, patientID uuid.UUID, req *BookAppointmentRequest) (*Appointment, error)
	GetAppointments(ctx context.Context, currentUserID uuid.UUID, role Role, filter AppointmentFilter) ([]Appointment, int64, error)
	GetAppointmentByID(ctx context.Context, currentUserID uuid.UUID, role Role, id uuid.UUID) (*Appointment, error)
	ConfirmAppointment(ctx context.Context, currentUserID uuid.UUID, role Role, id uuid.UUID) (*Appointment, error)
	CancelAppointment(ctx context.Context, currentUserID uuid.UUID, role Role, id uuid.UUID, req *CancelAppointmentRequest) (*Appointment, error)
	CompleteAppointment(ctx context.Context, currentUserID uuid.UUID, role Role, id uuid.UUID, req *CompleteAppointmentRequest) (*Appointment, error)
}
