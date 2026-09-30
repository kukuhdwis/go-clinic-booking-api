package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"projectgo/internal/delivery/http/middleware"
	"projectgo/internal/domain"
	"projectgo/internal/usecase"
	"projectgo/pkg/response"
)

type ScheduleHandler struct {
	scheduleUsecase domain.ScheduleUsecase
}

func NewScheduleHandler(scheduleUsecase domain.ScheduleUsecase) *ScheduleHandler {
	return &ScheduleHandler{scheduleUsecase: scheduleUsecase}
}

// CreateSchedule godoc
// @Summary Add working hours schedule for a doctor (Admin only)
// @Tags Schedules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateScheduleRequest true "Schedule details"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/admin/schedules [post]
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req domain.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	schedule, err := h.scheduleUsecase.CreateSchedule(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrInvalidTimeRange) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to create schedule: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Doctor schedule created successfully", schedule)
}

// GetDoctorSchedules godoc
// @Summary Get all recurring working schedules for a doctor
// @Tags Schedules
// @Produce json
// @Param id path string true "Doctor UUID"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/doctors/{id}/schedules [get]
func (h *ScheduleHandler) GetDoctorSchedules(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid doctor ID format", nil)
		return
	}

	schedules, err := h.scheduleUsecase.GetDoctorSchedules(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve schedules: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Doctor schedules retrieved successfully", schedules)
}

// GetAvailableSlots godoc
// @Summary Calculate and get available appointment booking slots for a doctor on a target date
// @Tags Schedules
// @Produce json
// @Param id path string true "Doctor UUID"
// @Param date query string true "Target date format YYYY-MM-DD"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/doctors/{id}/available-slots [get]
func (h *ScheduleHandler) GetAvailableSlots(c *gin.Context) {
	idStr := c.Param("id")
	doctorID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid doctor ID format", nil)
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		response.BadRequest(c, "Missing required query parameter: date (YYYY-MM-DD)", nil)
		return
	}

	availability, err := h.scheduleUsecase.GetAvailableSlots(c.Request.Context(), doctorID, dateStr)
	if err != nil {
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrInvalidDateFormat) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to calculate available slots: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Available slots retrieved successfully", availability)
}

// GetMySchedules godoc
// @Summary Get recurring working shifts for the authenticated doctor
// @Tags Schedules
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIErrorResponse
// @Router /api/v1/doctors/me/schedules [get]
func (h *ScheduleHandler) GetMySchedules(c *gin.Context) {
	doctorUserID := middleware.GetCurrentUserID(c)
	schedules, err := h.scheduleUsecase.GetDoctorOwnSchedules(c.Request.Context(), doctorUserID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve schedules: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Schedules retrieved successfully", schedules)
}

// CreateMySchedule godoc
// @Summary Add a working shift for the authenticated doctor
// @Tags Schedules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.DoctorCreateScheduleRequest true "Shift details"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Router /api/v1/doctors/me/schedules [post]
func (h *ScheduleHandler) CreateMySchedule(c *gin.Context) {
	doctorUserID := middleware.GetCurrentUserID(c)

	var req domain.DoctorCreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	schedule, err := h.scheduleUsecase.CreateDoctorOwnSchedule(c.Request.Context(), doctorUserID, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidTimeRange) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create schedule: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Working schedule created successfully", schedule)
}

// DeleteMySchedule godoc
// @Summary Delete a working shift for the authenticated doctor
// @Tags Schedules
// @Security BearerAuth
// @Produce json
// @Param id path string true "Schedule UUID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/doctors/me/schedules/{id} [delete]
func (h *ScheduleHandler) DeleteMySchedule(c *gin.Context) {
	doctorUserID := middleware.GetCurrentUserID(c)
	scheduleIDStr := c.Param("id")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid schedule ID format", nil)
		return
	}

	if err := h.scheduleUsecase.DeleteDoctorSchedule(c.Request.Context(), doctorUserID, scheduleID); err != nil {
		if errors.Is(err, usecase.ErrDoctorScheduleNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to delete schedule: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Working schedule removed successfully", nil)
}
