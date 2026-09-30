package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"projectgo/internal/domain"
	"projectgo/pkg/password"
)

var (
	ErrDoctorNotFound      = errors.New("doctor not found")
	ErrDoctorAlreadyExists  = errors.New("doctor profile already exists for this user")
	ErrInvalidDoctorInput  = errors.New("either user_id or complete doctor credentials (email, password, full_name) must be provided")
)

type doctorUsecase struct {
	doctorRepo domain.DoctorRepository
	userRepo   domain.UserRepository
	clinicRepo domain.ClinicRepository
}

func NewDoctorUsecase(doctorRepo domain.DoctorRepository, userRepo domain.UserRepository, clinicRepo domain.ClinicRepository) domain.DoctorUsecase {
	return &doctorUsecase{
		doctorRepo: doctorRepo,
		userRepo:   userRepo,
		clinicRepo: clinicRepo,
	}
}

func (u *doctorUsecase) CreateDoctor(ctx context.Context, req *domain.CreateDoctorRequest) (*domain.Doctor, error) {
	// 1. Verify target clinic exists
	_, err := u.clinicRepo.FindByID(ctx, req.ClinicID)
	if err != nil {
		return nil, ErrClinicNotFound
	}

	var targetUserID uuid.UUID

	// 2. Determine if linking existing user or creating new doctor user account atomically
	if req.UserID != uuid.Nil {
		// Option A: Link existing user
		user, err := u.userRepo.FindByID(ctx, req.UserID)
		if err != nil {
			return nil, ErrUserNotFound
		}
		targetUserID = user.ID

		// Ensure user's role is set to doctor
		if user.Role != domain.RoleDoctor && user.Role != domain.RoleAdmin {
			user.Role = domain.RoleDoctor
			_ = u.userRepo.Create(ctx, user)
		}
	} else {
		// Option B: Atomic creation of user account with role doctor
		if req.Email == "" || req.Password == "" || req.FullName == "" {
			return nil, ErrInvalidDoctorInput
		}

		// Ensure email is not already registered
		existingUser, _ := u.userRepo.FindByEmail(ctx, req.Email)
		if existingUser != nil {
			return nil, ErrEmailAlreadyExists
		}

		// Hash password
		hashedPassword, err := password.Hash(req.Password)
		if err != nil {
			return nil, err
		}

		newUser := &domain.User{
			Email:        req.Email,
			PasswordHash: hashedPassword,
			FullName:     req.FullName,
			Phone:        req.Phone,
			Role:         domain.RoleDoctor,
		}

		if err := u.userRepo.Create(ctx, newUser); err != nil {
			return nil, err
		}

		targetUserID = newUser.ID
	}

	// 3. Check if doctor profile already created for this user
	existing, _ := u.doctorRepo.FindByUserID(ctx, targetUserID)
	if existing != nil {
		return nil, ErrDoctorAlreadyExists
	}

	slotDuration := req.SlotDurationMinutes
	if slotDuration <= 0 {
		slotDuration = 30
	}

	doctor := &domain.Doctor{
		UserID:              targetUserID,
		ClinicID:            req.ClinicID,
		Specialization:      req.Specialization,
		SlotDurationMinutes: slotDuration,
		ConsultationFee:     req.ConsultationFee,
		Bio:                 req.Bio,
	}

	if err := u.doctorRepo.Create(ctx, doctor); err != nil {
		return nil, err
	}

	return u.doctorRepo.FindByID(ctx, doctor.ID)
}

func (u *doctorUsecase) GetDoctorByID(ctx context.Context, id uuid.UUID) (*domain.Doctor, error) {
	doctor, err := u.doctorRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrDoctorNotFound
	}
	return doctor, nil
}

func (u *doctorUsecase) GetDoctorByUserID(ctx context.Context, userID uuid.UUID) (*domain.Doctor, error) {
	doctor, err := u.doctorRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, ErrDoctorNotFound
	}
	return doctor, nil
}

func (u *doctorUsecase) UpdateDoctorProfile(ctx context.Context, userID uuid.UUID, req *domain.UpdateDoctorProfileRequest) (*domain.Doctor, error) {
	doctor, err := u.doctorRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, ErrDoctorNotFound
	}

	if req.Specialization != nil && *req.Specialization != "" {
		doctor.Specialization = *req.Specialization
	}
	if req.SlotDurationMinutes != nil && *req.SlotDurationMinutes > 0 {
		doctor.SlotDurationMinutes = *req.SlotDurationMinutes
	}
	if req.ConsultationFee != nil && *req.ConsultationFee >= 0 {
		doctor.ConsultationFee = *req.ConsultationFee
	}
	if req.Bio != nil {
		doctor.Bio = *req.Bio
	}

	if err := u.doctorRepo.Update(ctx, doctor); err != nil {
		return nil, err
	}

	return u.doctorRepo.FindByID(ctx, doctor.ID)
}

func (u *doctorUsecase) GetDoctors(ctx context.Context, filter domain.DoctorFilter) ([]domain.Doctor, int64, error) {
	return u.doctorRepo.FindAll(ctx, filter)
}
