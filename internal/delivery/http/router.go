package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "projectgo/docs"
	"projectgo/internal/delivery/http/handler"
	"projectgo/internal/delivery/http/middleware"
	"projectgo/internal/domain"
	"projectgo/pkg/response"
	"projectgo/pkg/token"
)

type RouterConfig struct {
	AuthHandler        *handler.AuthHandler
	ClinicHandler      *handler.ClinicHandler
	DoctorHandler      *handler.DoctorHandler
	ScheduleHandler    *handler.ScheduleHandler
	AppointmentHandler *handler.AppointmentHandler
	JWTManager         *token.JWTManager
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "ClinicEase API is healthy and operational", gin.H{
			"status": "UP",
		})
	})

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve Web Dashboard Frontend
	r.Static("/static", "./web")
	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})
	r.GET("/dashboard", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	apiV1 := r.Group("/api/v1")
	{
		// 1. Auth routes (Public)
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/register", cfg.AuthHandler.Register)
			authRoutes.POST("/login", cfg.AuthHandler.Login)

			// Protected profile
			authRoutes.GET("/me", middleware.AuthMiddleware(cfg.JWTManager), cfg.AuthHandler.GetProfile)
		}

		// 2. Clinics (Public read, Admin write)
		clinicRoutes := apiV1.Group("/clinics")
		{
			clinicRoutes.GET("", cfg.ClinicHandler.GetClinics)
			clinicRoutes.GET("/:id", cfg.ClinicHandler.GetClinicByID)
		}

		// 3. Doctors (Public read, Doctor self-service, Admin write)
		doctorRoutes := apiV1.Group("/doctors")
		{
			// Doctor self-service (registered before :id parameter)
			doctorMeRoutes := doctorRoutes.Group("/me")
			doctorMeRoutes.Use(middleware.AuthMiddleware(cfg.JWTManager), middleware.RequireRole(domain.RoleDoctor))
			{
				doctorMeRoutes.GET("", cfg.DoctorHandler.GetMyProfile)
				doctorMeRoutes.PATCH("", cfg.DoctorHandler.UpdateMyProfile)
				doctorMeRoutes.GET("/schedules", cfg.ScheduleHandler.GetMySchedules)
				doctorMeRoutes.POST("/schedules", cfg.ScheduleHandler.CreateMySchedule)
				doctorMeRoutes.DELETE("/schedules/:id", cfg.ScheduleHandler.DeleteMySchedule)
			}

			doctorRoutes.GET("", cfg.DoctorHandler.GetDoctors)
			doctorRoutes.GET("/:id", cfg.DoctorHandler.GetDoctorByID)
			doctorRoutes.GET("/:id/schedules", cfg.ScheduleHandler.GetDoctorSchedules)
			doctorRoutes.GET("/:id/available-slots", cfg.ScheduleHandler.GetAvailableSlots)
		}

		// 4. Appointments (Protected)
		appointmentRoutes := apiV1.Group("/appointments")
		appointmentRoutes.Use(middleware.AuthMiddleware(cfg.JWTManager))
		{
			// Patient book appointment
			appointmentRoutes.POST("", middleware.RequireRole(domain.RolePatient), cfg.AppointmentHandler.BookAppointment)

			// View appointment history (Patient / Doctor / Admin)
			appointmentRoutes.GET("/my", cfg.AppointmentHandler.GetMyAppointments)
			appointmentRoutes.GET("/:id", cfg.AppointmentHandler.GetAppointmentByID)

			// Status updates
			appointmentRoutes.PATCH("/:id/confirm", middleware.RequireRole(domain.RoleDoctor, domain.RoleAdmin), cfg.AppointmentHandler.ConfirmAppointment)
			appointmentRoutes.PATCH("/:id/cancel", cfg.AppointmentHandler.CancelAppointment)
			appointmentRoutes.POST("/:id/complete", middleware.RequireRole(domain.RoleDoctor), cfg.AppointmentHandler.CompleteAppointment)
		}

		// 5. Admin management routes
		adminRoutes := apiV1.Group("/admin")
		adminRoutes.Use(middleware.AuthMiddleware(cfg.JWTManager), middleware.RequireRole(domain.RoleAdmin))
		{
			adminRoutes.POST("/clinics", cfg.ClinicHandler.CreateClinic)
			adminRoutes.POST("/doctors", cfg.DoctorHandler.CreateDoctor)
			adminRoutes.POST("/schedules", cfg.ScheduleHandler.CreateSchedule)
		}
	}

	return r
}
