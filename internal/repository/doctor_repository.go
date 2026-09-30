package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"projectgo/internal/domain"
)

type doctorRepository struct {
	db *gorm.DB
}

func NewDoctorRepository(db *gorm.DB) domain.DoctorRepository {
	return &doctorRepository{db: db}
}

func (r *doctorRepository) Create(ctx context.Context, doctor *domain.Doctor) error {
	return r.db.WithContext(ctx).Create(doctor).Error
}

func (r *doctorRepository) Update(ctx context.Context, doctor *domain.Doctor) error {
	return r.db.WithContext(ctx).Save(doctor).Error
}

func (r *doctorRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Doctor, error) {
	var doctor domain.Doctor
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Clinic").
		Preload("Schedules", "is_active = ?", true).
		Where("id = ?", id).
		First(&doctor).Error
	if err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *doctorRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.Doctor, error) {
	var doctor domain.Doctor
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Clinic").
		Where("user_id = ?", userID).
		First(&doctor).Error
	if err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *doctorRepository) FindAll(ctx context.Context, filter domain.DoctorFilter) ([]domain.Doctor, int64, error) {
	var doctors []domain.Doctor
	var totalRecords int64

	query := r.db.WithContext(ctx).Model(&domain.Doctor{}).
		Joins("JOIN users ON users.id = doctors.user_id AND users.deleted_at IS NULL").
		Joins("JOIN clinics ON clinics.id = doctors.clinic_id AND clinics.deleted_at IS NULL")

	if filter.Specialization != "" {
		query = query.Where("doctors.specialization LIKE ?", "%"+filter.Specialization+"%")
	}

	if filter.ClinicID != uuid.Nil {
		query = query.Where("doctors.clinic_id = ?", filter.ClinicID)
	}

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Where("users.full_name LIKE ? OR doctors.specialization LIKE ?", searchTerm, searchTerm)
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
		Preload("User").
		Preload("Clinic").
		Offset(offset).
		Limit(limit).
		Find(&doctors).Error

	if err != nil {
		return nil, 0, err
	}

	return doctors, totalRecords, nil
}
