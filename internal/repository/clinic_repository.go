package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"projectgo/internal/domain"
)

type clinicRepository struct {
	db *gorm.DB
}

func NewClinicRepository(db *gorm.DB) domain.ClinicRepository {
	return &clinicRepository{db: db}
}

func (r *clinicRepository) Create(ctx context.Context, clinic *domain.Clinic) error {
	return r.db.WithContext(ctx).Create(clinic).Error
}

func (r *clinicRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Clinic, error) {
	var clinic domain.Clinic
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&clinic).Error; err != nil {
		return nil, err
	}
	return &clinic, nil
}

func (r *clinicRepository) FindAll(ctx context.Context) ([]domain.Clinic, error) {
	var clinics []domain.Clinic
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&clinics).Error; err != nil {
		return nil, err
	}
	return clinics, nil
}
