package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"projectgo/internal/domain"
)

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) domain.ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(ctx context.Context, schedule *domain.DoctorSchedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *scheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.DoctorSchedule, error) {
	var schedule domain.DoctorSchedule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *scheduleRepository) FindByDoctorAndDay(ctx context.Context, doctorID uuid.UUID, dayOfWeek int) ([]domain.DoctorSchedule, error) {
	var schedules []domain.DoctorSchedule
	err := r.db.WithContext(ctx).
		Where("doctor_id = ? AND day_of_week = ? AND is_active = ?", doctorID, dayOfWeek, true).
		Order("start_time ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *scheduleRepository) FindByDoctorID(ctx context.Context, doctorID uuid.UUID) ([]domain.DoctorSchedule, error) {
	var schedules []domain.DoctorSchedule
	err := r.db.WithContext(ctx).
		Where("doctor_id = ?", doctorID).
		Order("day_of_week ASC, start_time ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *scheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.DoctorSchedule{}).Error
}
