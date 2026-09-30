package handler

import (
	"errors"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"projectgo/internal/delivery/http/middleware"
	"projectgo/internal/domain"
	"projectgo/internal/usecase"
	"projectgo/pkg/response"
)

type AppointmentHandler struct {
	appointmentUsecase domain.AppointmentUsecase
}

func NewAppointmentHandler(appointmentUsecase domain.AppointmentUsecase) *AppointmentHandler {
	return &AppointmentHandler{appointmentUsecase: appointmentUsecase}
}

// BookAppointment godoc
// @Summary Book an appointment slot with atomic transaction locking (Patient only)
// @Tags Appointments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.BookAppointmentRequest true "Booking payload"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 409 {object} response.APIErrorResponse
// @Router /api/v1/appointments [post]
func (h *AppointmentHandler) BookAppointment(c *gin.Context) {
	patientID := middleware.GetCurrentUserID(c)

	var req domain.BookAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	appointment, err := h.appointmentUsecase.BookAppointment(c.Request.Context(), patientID, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrAppointmentConflict) {
			response.Conflict(c, err.Error(), "SLOT_ALREADY_BOOKED")
			return
		}
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrAppointmentInPast) ||
			errors.Is(err, usecase.ErrDoctorNotAvailableOnSlot) ||
			errors.Is(err, usecase.ErrInvalidDateFormat) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to book appointment: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Appointment booked successfully", appointment)
}

// GetMyAppointments godoc
// @Summary List current user's appointments (Scoped by role with filters & pagination)
// @Tags Appointments
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status: pending, confirmed, completed, cancelled"
// @Param date query string false "Filter by specific date YYYY-MM-DD"
// @Param start_date query string false "Filter start date range YYYY-MM-DD"
// @Param end_date query string false "Filter end date range YYYY-MM-DD"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/appointments/my [get]
func (h *AppointmentHandler) GetMyAppointments(c *gin.Context) {
	currentUserID := middleware.GetCurrentUserID(c)
	role := middleware.GetCurrentUserRole(c)

	var filter domain.AppointmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	appointments, total, err := h.appointmentUsecase.GetAppointments(c.Request.Context(), currentUserID, role, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve appointments: "+err.Error())
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	meta := response.Pagination{
		Page:         filter.Page,
		Limit:        filter.Limit,
		TotalRecords: total,
		TotalPages:   totalPages,
	}

	response.SuccessWithPagination(c, http.StatusOK, "Appointments retrieved successfully", appointments, meta)
}

// GetAppointmentByID godoc
// @Summary Get appointment details by ID
// @Tags Appointments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Appointment UUID"
// @Success 200 {object} response.APIResponse
// @Failure 403 {object} response.APIErrorResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/appointments/{id} [get]
func (h *AppointmentHandler) GetAppointmentByID(c *gin.Context) {
	currentUserID := middleware.GetCurrentUserID(c)
	role := middleware.GetCurrentUserRole(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID format", nil)
		return
	}

	appointment, err := h.appointmentUsecase.GetAppointmentByID(c.Request.Context(), currentUserID, role, id)
	if err != nil {
		if errors.Is(err, usecase.ErrUnauthorizedAppointment) {
			response.Forbidden(c, err.Error())
			return
		}
		response.NotFound(c, "Appointment not found")
		return
	}

	response.Success(c, http.StatusOK, "Appointment retrieved successfully", appointment)
}

// ConfirmAppointment godoc
// @Summary Confirm a pending appointment (Doctor or Admin)
// @Tags Appointments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Appointment UUID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 403 {object} response.APIErrorResponse
// @Router /api/v1/appointments/{id}/confirm [patch]
func (h *AppointmentHandler) ConfirmAppointment(c *gin.Context) {
	currentUserID := middleware.GetCurrentUserID(c)
	role := middleware.GetCurrentUserRole(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID format", nil)
		return
	}

	appointment, err := h.appointmentUsecase.ConfirmAppointment(c.Request.Context(), currentUserID, role, id)
	if err != nil {
		if errors.Is(err, usecase.ErrUnauthorizedAppointment) {
			response.Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrInvalidStatusTransition) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to confirm appointment: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Appointment confirmed successfully", appointment)
}

// CancelAppointment godoc
// @Summary Cancel an appointment
// @Tags Appointments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Appointment UUID"
// @Param request body domain.CancelAppointmentRequest true "Cancellation reason"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 403 {object} response.APIErrorResponse
// @Router /api/v1/appointments/{id}/cancel [patch]
func (h *AppointmentHandler) CancelAppointment(c *gin.Context) {
	currentUserID := middleware.GetCurrentUserID(c)
	role := middleware.GetCurrentUserRole(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID format", nil)
		return
	}

	var req domain.CancelAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Cancellation reason is required", err.Error())
		return
	}

	appointment, err := h.appointmentUsecase.CancelAppointment(c.Request.Context(), currentUserID, role, id, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrUnauthorizedAppointment) {
			response.Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrCancellationDeadlinePassed) || errors.Is(err, usecase.ErrInvalidStatusTransition) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to cancel appointment: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Appointment cancelled successfully", appointment)
}

// CompleteAppointment godoc
// @Summary Complete consultation and submit medical diagnosis notes (Doctor only)
// @Tags Appointments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Appointment UUID"
// @Param request body domain.CompleteAppointmentRequest true "Diagnosis and prescription"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 403 {object} response.APIErrorResponse
// @Router /api/v1/appointments/{id}/complete [post]
func (h *AppointmentHandler) CompleteAppointment(c *gin.Context) {
	currentUserID := middleware.GetCurrentUserID(c)
	role := middleware.GetCurrentUserRole(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID format", nil)
		return
	}

	var req domain.CompleteAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid medical notes payload", err.Error())
		return
	}

	appointment, err := h.appointmentUsecase.CompleteAppointment(c.Request.Context(), currentUserID, role, id, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrUnauthorizedAppointment) {
			response.Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrInvalidStatusTransition) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to complete appointment: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Appointment completed and medical notes saved successfully", appointment)
}
