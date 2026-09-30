package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"projectgo/internal/domain"
)

var (
	ErrClinicNotFound = errors.New("clinic not found")
)

type clinicUsecase struct {
	clinicRepo domain.ClinicRepository
}

func NewClinicUsecase(clinicRepo domain.ClinicRepository) domain.ClinicUsecase {
	return &clinicUsecase{clinicRepo: clinicRepo}
}

func (u *clinicUsecase) CreateClinic(ctx context.Context, req *domain.CreateClinicRequest) (*domain.Clinic, error) {
	clinic := &domain.Clinic{
		Name:    req.Name,
		Address: req.Address,
		Phone:   req.Phone,
	}

	if err := u.clinicRepo.Create(ctx, clinic); err != nil {
		return nil, err
	}

	return clinic, nil
}

func (u *clinicUsecase) GetClinics(ctx context.Context) ([]domain.Clinic, error) {
	return u.clinicRepo.FindAll(ctx)
}

func (u *clinicUsecase) GetClinicByID(ctx context.Context, id uuid.UUID) (*domain.Clinic, error) {
	clinic, err := u.clinicRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrClinicNotFound
	}
	return clinic, nil
}
