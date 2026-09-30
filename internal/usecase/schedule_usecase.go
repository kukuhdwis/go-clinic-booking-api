package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"projectgo/internal/domain"
)

var (
	ErrInvalidTimeRange = errors.New("start time must be earlier than end time")
	ErrDoctorScheduleNotFound = errors.New("no schedules found for this doctor")
	ErrInvalidDateFormat = errors.New("invalid date format, must be YYYY-MM-DD")
)

type scheduleUsecase struct {
	scheduleRepo    domain.ScheduleRepository
	doctorRepo      domain.DoctorRepository
	appointmentRepo domain.AppointmentRepository
}

func NewScheduleUsecase(
	scheduleRepo domain.ScheduleRepository,
	doctorRepo domain.DoctorRepository,
	appointmentRepo domain.AppointmentRepository,
) domain.ScheduleUsecase {
	return &scheduleUsecase{
		scheduleRepo:    scheduleRepo,
		doctorRepo:      doctorRepo,
		appointmentRepo: appointmentRepo,
	}
}

func (u *scheduleUsecase) CreateSchedule(ctx context.Context, req *domain.CreateScheduleRequest) (*domain.DoctorSchedule, error) {
	// Verify doctor exists
	_, err := u.doctorRepo.FindByID(ctx, req.DoctorID)
	if err != nil {
		return nil, ErrDoctorNotFound
	}

	// Validate start_time < end_time
	startT, err1 := parseTime(req.StartTime)
	endT, err2 := parseTime(req.EndTime)
	if err1 != nil || err2 != nil || !startT.Before(endT) {
		return nil, ErrInvalidTimeRange
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	schedule := &domain.DoctorSchedule{
		DoctorID:  req.DoctorID,
		DayOfWeek: req.DayOfWeek,
		StartTime: formatTime(startT),
		EndTime:   formatTime(endT),
		IsActive:  isActive,
	}

	if err := u.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (u *scheduleUsecase) GetDoctorSchedules(ctx context.Context, doctorID uuid.UUID) ([]domain.DoctorSchedule, error) {
	return u.scheduleRepo.FindByDoctorID(ctx, doctorID)
}

func (u *scheduleUsecase) CreateDoctorOwnSchedule(ctx context.Context, doctorUserID uuid.UUID, req *domain.DoctorCreateScheduleRequest) (*domain.DoctorSchedule, error) {
	doctor, err := u.doctorRepo.FindByUserID(ctx, doctorUserID)
	if err != nil {
		return nil, ErrDoctorNotFound
	}

	return u.CreateSchedule(ctx, &domain.CreateScheduleRequest{
		DoctorID:  doctor.ID,
		DayOfWeek: req.DayOfWeek,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		IsActive:  req.IsActive,
	})
}

func (u *scheduleUsecase) GetDoctorOwnSchedules(ctx context.Context, doctorUserID uuid.UUID) ([]domain.DoctorSchedule, error) {
	doctor, err := u.doctorRepo.FindByUserID(ctx, doctorUserID)
	if err != nil {
		return nil, ErrDoctorNotFound
	}

	return u.GetDoctorSchedules(ctx, doctor.ID)
}

func (u *scheduleUsecase) DeleteDoctorSchedule(ctx context.Context, doctorUserID uuid.UUID, scheduleID uuid.UUID) error {
	doctor, err := u.doctorRepo.FindByUserID(ctx, doctorUserID)
	if err != nil {
		return ErrDoctorNotFound
	}

	schedule, err := u.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return ErrDoctorScheduleNotFound
	}

	if schedule.DoctorID != doctor.ID {
		return errors.New("unauthorized to delete this schedule")
	}

	return u.scheduleRepo.Delete(ctx, scheduleID)
}

func (u *scheduleUsecase) GetAvailableSlots(ctx context.Context, doctorID uuid.UUID, targetDateStr string) (*domain.DoctorAvailabilityResponse, error) {
	targetDate, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	doctor, err := u.doctorRepo.FindByID(ctx, doctorID)
	if err != nil {
		return nil, ErrDoctorNotFound
	}

	dayOfWeek := int(targetDate.Weekday()) // 0=Sunday, 1=Monday...
	schedules, err := u.scheduleRepo.FindByDoctorAndDay(ctx, doctorID, dayOfWeek)
	if err != nil {
		return nil, err
	}

	// Fetch active appointments booked for this doctor on targetDate
	activeAppointments, err := u.appointmentRepo.FindActiveByDoctorAndDate(ctx, doctorID, targetDateStr)
	if err != nil {
		return nil, err
	}

	bookedSlots := make(map[string]bool)
	for _, appt := range activeAppointments {
		bookedSlots[appt.StartTime] = true
	}

	slotDuration := doctor.SlotDurationMinutes
	if slotDuration <= 0 {
		slotDuration = 30
	}

	var allSlots []domain.AvailableSlot
	now := time.Now()
	isToday := targetDate.Format("2006-01-02") == now.Format("2006-01-02")

	for _, sch := range schedules {
		startTime, err1 := parseTime(sch.StartTime)
		endTime, err2 := parseTime(sch.EndTime)
		if err1 != nil || err2 != nil {
			continue
		}

		current := startTime
		for {
			next := current.Add(time.Duration(slotDuration) * time.Minute)
			if next.After(endTime) {
				break
			}

			startStr := formatTime(current)
			endStr := formatTime(next)

			isAvailable := true
			// Check if already booked
			if bookedSlots[startStr] {
				isAvailable = false
			}

			// If appointment date is today, check if slot is already in the past
			if isToday {
				slotHour, slotMin := current.Hour(), current.Minute()
				if now.Hour() > slotHour || (now.Hour() == slotHour && now.Minute() >= slotMin) {
					isAvailable = false
				}
			}

			allSlots = append(allSlots, domain.AvailableSlot{
				StartTime:   startStr,
				EndTime:     endStr,
				IsAvailable: isAvailable,
			})

			current = next
		}
	}

	doctorName := ""
	if doctor.User.FullName != "" {
		doctorName = doctor.User.FullName
	}

	return &domain.DoctorAvailabilityResponse{
		DoctorID:       doctor.ID,
		DoctorName:     doctorName,
		Specialization: doctor.Specialization,
		Date:           targetDateStr,
		DayOfWeek:      dayOfWeek,
		SlotDuration:   slotDuration,
		Slots:          allSlots,
	}, nil
}

func parseTime(val string) (time.Time, error) {
	if len(val) == 5 { // "15:04"
		return time.Parse("15:04", val)
	}
	return time.Parse("15:04:05", val)
}

func formatTime(t time.Time) string {
	return fmt.Sprintf("%02d:%02d:00", t.Hour(), t.Minute())
}
