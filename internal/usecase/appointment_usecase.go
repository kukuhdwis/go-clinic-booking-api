package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"projectgo/internal/domain"
	"projectgo/internal/repository"
)

var (
	ErrAppointmentConflict        = errors.New("the selected appointment slot is already booked")
	ErrUnauthorizedAppointment    = errors.New("you are not authorized to view or modify this appointment")
	ErrInvalidStatusTransition    = errors.New("invalid appointment status transition")
	ErrCancellationDeadlinePassed = errors.New("appointments can only be cancelled at least 2 hours in advance")
	ErrAppointmentInPast          = errors.New("cannot book an appointment in the past")
	ErrDoctorNotAvailableOnSlot   = errors.New("doctor has no active schedule for the selected time slot")
)

type appointmentUsecase struct {
	appointmentRepo domain.AppointmentRepository
	doctorRepo      domain.DoctorRepository
	scheduleRepo    domain.ScheduleRepository
	userRepo        domain.UserRepository
}

func NewAppointmentUsecase(
	appointmentRepo domain.AppointmentRepository,
	doctorRepo domain.DoctorRepository,
	scheduleRepo domain.ScheduleRepository,
	userRepo domain.UserRepository,
) domain.AppointmentUsecase {
	return &appointmentUsecase{
		appointmentRepo: appointmentRepo,
		doctorRepo:      doctorRepo,
		scheduleRepo:    scheduleRepo,
		userRepo:        userRepo,
	}
}

func (u *appointmentUsecase) BookAppointment(ctx context.Context, patientID uuid.UUID, req *domain.BookAppointmentRequest) (*domain.Appointment, error) {
	// Parse appointment date
	parsedDate, err := time.Parse("2006-01-02", req.AppointmentDate)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	// Parse start time
	startTime, err := parseTime(req.StartTime)
	if err != nil {
		return nil, errors.New("invalid start_time format (expected HH:MM)")
	}

	// Verify not booking in past
	now := time.Now()
	appointmentDateTime := time.Date(
		parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.Local,
	)
	if appointmentDateTime.Before(now) {
		return nil, ErrAppointmentInPast
	}

	// Verify doctor exists
	doctor, err := u.doctorRepo.FindByID(ctx, req.DoctorID)
	if err != nil {
		return nil, ErrDoctorNotFound
	}

	// Verify doctor works on this day & time
	dayOfWeek := int(parsedDate.Weekday())
	schedules, err := u.scheduleRepo.FindByDoctorAndDay(ctx, req.DoctorID, dayOfWeek)
	if err != nil || len(schedules) == 0 {
		return nil, ErrDoctorNotAvailableOnSlot
	}

	slotDuration := doctor.SlotDurationMinutes
	if slotDuration <= 0 {
		slotDuration = 30
	}
	endTime := startTime.Add(time.Duration(slotDuration) * time.Minute)

	startTimeStr := formatTime(startTime)
	endTimeStr := formatTime(endTime)

	// Check if slot falls within any doctor schedule block
	slotValid := false
	for _, sch := range schedules {
		schStart, _ := parseTime(sch.StartTime)
		schEnd, _ := parseTime(sch.EndTime)
		if (startTime.Equal(schStart) || startTime.After(schStart)) &&
			(endTime.Equal(schEnd) || endTime.Before(schEnd)) {
			slotValid = true
			break
		}
	}

	if !slotValid {
		return nil, ErrDoctorNotAvailableOnSlot
	}

	appointment := &domain.Appointment{
		PatientID:            patientID,
		DoctorID:             req.DoctorID,
		AppointmentDate:      req.AppointmentDate,
		StartTime:            startTimeStr,
		EndTime:              endTimeStr,
		Status:               domain.StatusPending,
		ComplaintDescription: req.ComplaintDescription,
	}

	// Atomic insert with lock
	if err := u.appointmentRepo.CreateWithLock(ctx, appointment); err != nil {
		if errors.Is(err, repository.ErrSlotAlreadyBooked) {
			return nil, ErrAppointmentConflict
		}
		return nil, err
	}

	return u.appointmentRepo.FindByID(ctx, appointment.ID)
}

func (u *appointmentUsecase) GetAppointments(ctx context.Context, currentUserID uuid.UUID, role domain.Role, filter domain.AppointmentFilter) ([]domain.Appointment, int64, error) {
	switch role {
	case domain.RolePatient:
		filter.PatientID = currentUserID
	case domain.RoleDoctor:
		doctor, err := u.doctorRepo.FindByUserID(ctx, currentUserID)
		if err != nil {
			return nil, 0, ErrDoctorNotFound
		}
		filter.DoctorID = doctor.ID
	case domain.RoleAdmin:
		// Admin can view all, or use provided filters
	default:
		return nil, 0, ErrUnauthorizedAppointment
	}

	return u.appointmentRepo.FindAll(ctx, filter)
}

func (u *appointmentUsecase) GetAppointmentByID(ctx context.Context, currentUserID uuid.UUID, role domain.Role, id uuid.UUID) (*domain.Appointment, error) {
	appointment, err := u.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := u.validateOwnership(currentUserID, role, appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

func (u *appointmentUsecase) ConfirmAppointment(ctx context.Context, currentUserID uuid.UUID, role domain.Role, id uuid.UUID) (*domain.Appointment, error) {
	appointment, err := u.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only assigned doctor or admin can confirm
	if role != domain.RoleAdmin {
		if role != domain.RoleDoctor || appointment.Doctor.UserID != currentUserID {
			return nil, ErrUnauthorizedAppointment
		}
	}

	if appointment.Status != domain.StatusPending {
		return nil, fmt.Errorf("%w: only pending appointments can be confirmed", ErrInvalidStatusTransition)
	}

	if err := u.appointmentRepo.UpdateStatus(ctx, id, domain.StatusConfirmed, ""); err != nil {
		return nil, err
	}

	return u.appointmentRepo.FindByID(ctx, id)
}

func (u *appointmentUsecase) CancelAppointment(ctx context.Context, currentUserID uuid.UUID, role domain.Role, id uuid.UUID, req *domain.CancelAppointmentRequest) (*domain.Appointment, error) {
	appointment, err := u.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := u.validateOwnership(currentUserID, role, appointment); err != nil {
		return nil, err
	}

	if appointment.Status == domain.StatusCompleted || appointment.Status == domain.StatusCancelled {
		return nil, fmt.Errorf("%w: cannot cancel an appointment that is already %s", ErrInvalidStatusTransition, appointment.Status)
	}

	// If cancelled by patient, check 2-hour rule
	if role == domain.RolePatient {
		parsedDate, _ := time.Parse("2006-01-02", appointment.AppointmentDate)
		startTime, _ := parseTime(appointment.StartTime)
		appointmentDateTime := time.Date(
			parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, time.Local,
		)

		if time.Until(appointmentDateTime) < 2*time.Hour {
			return nil, ErrCancellationDeadlinePassed
		}
	}

	if err := u.appointmentRepo.UpdateStatus(ctx, id, domain.StatusCancelled, req.Reason); err != nil {
		return nil, err
	}

	return u.appointmentRepo.FindByID(ctx, id)
}

func (u *appointmentUsecase) CompleteAppointment(ctx context.Context, currentUserID uuid.UUID, role domain.Role, id uuid.UUID, req *domain.CompleteAppointmentRequest) (*domain.Appointment, error) {
	appointment, err := u.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only assigned doctor can complete consultation
	if role != domain.RoleDoctor || appointment.Doctor.UserID != currentUserID {
		return nil, ErrUnauthorizedAppointment
	}

	if appointment.Status != domain.StatusConfirmed {
		return nil, fmt.Errorf("%w: only confirmed appointments can be marked as completed", ErrInvalidStatusTransition)
	}

	medicalNote := &domain.MedicalNote{
		AppointmentID: id,
		Diagnosis:     req.Diagnosis,
		Prescription:  req.Prescription,
		Notes:         req.Notes,
	}

	if err := u.appointmentRepo.CompleteWithMedicalNote(ctx, id, medicalNote); err != nil {
		return nil, err
	}

	return u.appointmentRepo.FindByID(ctx, id)
}

func (u *appointmentUsecase) validateOwnership(currentUserID uuid.UUID, role domain.Role, appt *domain.Appointment) error {
	if role == domain.RoleAdmin {
		return nil
	}
	if role == domain.RolePatient && appt.PatientID == currentUserID {
		return nil
	}
	if role == domain.RoleDoctor && appt.Doctor.UserID == currentUserID {
		return nil
	}
	return ErrUnauthorizedAppointment
}
