package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"projectgo/internal/domain"
	"projectgo/internal/usecase"
	"projectgo/pkg/response"
)

type ClinicHandler struct {
	clinicUsecase domain.ClinicUsecase
}

func NewClinicHandler(clinicUsecase domain.ClinicUsecase) *ClinicHandler {
	return &ClinicHandler{clinicUsecase: clinicUsecase}
}

// CreateClinic godoc
// @Summary Create a new clinic (Admin only)
// @Tags Clinics
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateClinicRequest true "Clinic details"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Router /api/v1/admin/clinics [post]
func (h *ClinicHandler) CreateClinic(c *gin.Context) {
	var req domain.CreateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	clinic, err := h.clinicUsecase.CreateClinic(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to create clinic: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Clinic created successfully", clinic)
}

// GetClinics godoc
// @Summary List all clinics
// @Tags Clinics
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/v1/clinics [get]
func (h *ClinicHandler) GetClinics(c *gin.Context) {
	clinics, err := h.clinicUsecase.GetClinics(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve clinics: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Clinics retrieved successfully", clinics)
}

// GetClinicByID godoc
// @Summary Get clinic by ID
// @Tags Clinics
// @Produce json
// @Param id path string true "Clinic UUID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIErrorResponse
// @Router /api/v1/clinics/{id} [get]
func (h *ClinicHandler) GetClinicByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid clinic ID format", nil)
		return
	}

	clinic, err := h.clinicUsecase.GetClinicByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrClinicNotFound) {
			response.NotFound(c, "Clinic not found")
			return
		}
		response.InternalServerError(c, "Failed to retrieve clinic: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Clinic retrieved successfully", clinic)
}
