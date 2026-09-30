package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"projectgo/internal/domain"
)

var (
	ErrSlotAlreadyBooked  = errors.New("selected appointment slot is already booked")
	ErrAppointmentNotFound = errors.New("appointment not found")
)

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) domain.AppointmentRepository {
	return &appointmentRepository{db: db}
}

// CreateWithLock executes an atomic database transaction with pessimistic locking
// (SELECT ... FOR UPDATE) to eliminate any race condition / double-booking.
func (r *appointmentRepository) CreateWithLock(ctx context.Context, appointment *domain.Appointment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		// Query existing active bookings for this doctor and time slot with pessimistic lock
		err := tx.Model(&domain.Appointment{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("doctor_id = ? AND appointment_date = ? AND start_time = ? AND status IN (?, ?)",
				appointment.DoctorID,
				appointment.AppointmentDate,
				appointment.StartTime,
				domain.StatusPending,
				domain.StatusConfirmed,
			).
			Count(&count).Error

		if err != nil {
			return err
		}

		if count > 0 {
			return ErrSlotAlreadyBooked
		}

		if err := tx.Create(appointment).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *appointmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Appointment, error) {
	var appointment domain.Appointment
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Doctor.User").
		Preload("Doctor.Clinic").
		Preload("MedicalNote").
		Where("id = ?", id).
		First(&appointment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) FindAll(ctx context.Context, filter domain.AppointmentFilter) ([]domain.Appointment, int64, error) {
	var appointments []domain.Appointment
	var totalRecords int64

	query := r.db.WithContext(ctx).Model(&domain.Appointment{})

	if filter.PatientID != uuid.Nil {
		query = query.Where("patient_id = ?", filter.PatientID)
	}

	if filter.DoctorID != uuid.Nil {
		query = query.Where("doctor_id = ?", filter.DoctorID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Date != "" {
		query = query.Where("appointment_date = ?", filter.Date)
	}

	if filter.StartDate != "" {
		query = query.Where("appointment_date >= ?", filter.StartDate)
	}

	if filter.EndDate != "" {
		query = query.Where("appointment_date <= ?", filter.EndDate)
	}

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	err := query.
		Preload("Patient").
		Preload("Doctor.User").
		Preload("Doctor.Clinic").
		Preload("MedicalNote").
		Order("appointment_date DESC, start_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&appointments).Error

	if err != nil {
		return nil, 0, err
	}

	return appointments, totalRecords, nil
}

func (r *appointmentRepository) FindActiveByDoctorAndDate(ctx context.Context, doctorID uuid.UUID, date string) ([]domain.Appointment, error) {
	var appointments []domain.Appointment
	err := r.db.WithContext(ctx).
		Where("doctor_id = ? AND appointment_date = ? AND status IN (?, ?)",
			doctorID, date, domain.StatusPending, domain.StatusConfirmed).
		Find(&appointments).Error
	if err != nil {
		return nil, err
	}
	return appointments, nil
}

func (r *appointmentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AppointmentStatus, cancelReason string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if cancelReason != "" {
		updates["cancel_reason"] = cancelReason
	}
	return r.db.WithContext(ctx).Model(&domain.Appointment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *appointmentRepository) CompleteWithMedicalNote(ctx context.Context, appointmentID uuid.UUID, note *domain.MedicalNote) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Appointment{}).
			Where("id = ?", appointmentID).
			Update("status", domain.StatusCompleted).Error; err != nil {
			return err
		}

		note.AppointmentID = appointmentID
		if err := tx.Create(note).Error; err != nil {
			return err
		}

		return nil
	})
}
