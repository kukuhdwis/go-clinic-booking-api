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

type DoctorHandler struct {
	doctorUsecase domain.DoctorUsecase
}

func NewDoctorHandler(doctorUsecase domain.DoctorUsecase) *DoctorHandler {
	return &DoctorHandler{doctorUsecase: doctorUsecase}
}

// CreateDoctor godoc
// @Summary Create a doctor profile (Admin only)
// @Tags Doctors
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateDoctorRequest true "Doctor details"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 404 {object} response.APIErrorResponse
// @Failure 409 {object} response.APIErrorResponse
// @Router /api/v1/admin/doctors [post]
func (h *DoctorHandler) CreateDoctor(c *gin.Context) {
	var req domain.CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	doctor, err := h.doctorUsecase.CreateDoctor(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) || errors.Is(err, usecase.ErrClinicNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			response.Conflict(c, err.Error(), "EMAIL_ALREADY_EXISTS")
			return
		}
		if errors.Is(err, usecase.ErrDoctorAlreadyExists) {
			response.Conflict(c, err.Error(), "DOCTOR_EXISTS")
			return
		}
		if errors.Is(err, usecase.ErrInvalidDoctorInput) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		response.InternalServerError(c, "Failed to create doctor: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Doctor profile created successfully", doctor)
}

// GetDoctors godoc
// @Summary Search and filter doctors with pagination
// @Tags Doctors
// @Produce json
// @Param specialization query string false "Specialization filter"
// @Param clinic_id query string false "Clinic UUID"
// @Param search query string false "Search doctor name or specialization"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/doctors [get]
func (h *DoctorHandler) GetDoctors(c *gin.Context) {
	var filter domain.DoctorFilter
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

	doctors, total, err := h.doctorUsecase.GetDoctors(c.Request.Context(), filter)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve doctors: "+err.Error())
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	meta := response.Pagination{
		Page:         filter.Page,
		Limit:        filter.Limit,
		TotalRecords: total,
		TotalPages:   totalPages,
	}

	response.SuccessWithPagination(c, http.StatusOK, "Doctors retrieved successfully", doctors, meta)
}

// GetDoctorByID godoc
// @Summary Get doctor by ID with working schedules
// @Tags Doctors
// @Produce json
// @Param id path string true "Doctor UUID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/doctors/{id} [get]
func (h *DoctorHandler) GetDoctorByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid doctor ID format", nil)
		return
	}

	doctor, err := h.doctorUsecase.GetDoctorByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			response.NotFound(c, "Doctor not found")
			return
		}
		response.InternalServerError(c, "Failed to retrieve doctor: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Doctor retrieved successfully", doctor)
}

// GetMyProfile godoc
// @Summary Get authenticated doctor's profile
// @Tags Doctors
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIErrorResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/doctors/me [get]
func (h *DoctorHandler) GetMyProfile(c *gin.Context) {
	doctorUserID := middleware.GetCurrentUserID(c)
	doctor, err := h.doctorUsecase.GetDoctorByUserID(c.Request.Context(), doctorUserID)
	if err != nil {
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			response.NotFound(c, "Doctor profile not found for this user")
			return
		}
		response.InternalServerError(c, "Failed to retrieve doctor profile: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Doctor profile retrieved successfully", doctor)
}

// UpdateMyProfile godoc
// @Summary Update authenticated doctor's consultation fee, bio, duration, and specialty
// @Tags Doctors
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.UpdateDoctorProfileRequest true "Profile update fields"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 401 {object} response.APIErrorResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/doctors/me [patch]
func (h *DoctorHandler) UpdateMyProfile(c *gin.Context) {
	doctorUserID := middleware.GetCurrentUserID(c)

	var req domain.UpdateDoctorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	doctor, err := h.doctorUsecase.UpdateDoctorProfile(c.Request.Context(), doctorUserID, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrDoctorNotFound) {
			response.NotFound(c, "Doctor profile not found")
			return
		}
		response.InternalServerError(c, "Failed to update doctor profile: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Doctor profile updated successfully", doctor)
}
