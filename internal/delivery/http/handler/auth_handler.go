package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"projectgo/internal/delivery/http/middleware"
	"projectgo/internal/domain"
	"projectgo/internal/usecase"
	"projectgo/pkg/response"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(authUsecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// Register godoc
// @Summary Register a new patient
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Patient registration details"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 409 {object} response.APIErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	authResp, err := h.authUsecase.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			response.Conflict(c, err.Error(), "EMAIL_ALREADY_EXISTS")
			return
		}
		response.InternalServerError(c, "Failed to register user: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", authResp)
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login credentials"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIErrorResponse
// @Failure 401 {object} response.APIErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input payload", err.Error())
		return
	}

	authResp, err := h.authUsecase.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		response.InternalServerError(c, "Failed to login: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", authResp)
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	user, err := h.authUsecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "User profile not found")
		return
	}

	profile := domain.UserBrief{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", profile)
}
