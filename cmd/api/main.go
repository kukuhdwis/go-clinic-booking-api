package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
	"projectgo/internal/config"
	deliveryHttp "projectgo/internal/delivery/http"
	"projectgo/internal/delivery/http/handler"
	"projectgo/internal/domain"
	"projectgo/internal/repository"
	"projectgo/internal/usecase"
	"projectgo/pkg/password"
	"projectgo/pkg/token"
)

// @title ClinicEase API
// @version 1.0
// @description Production-grade Doctor Appointment & Clinic Booking API with Zero Double-Booking Guarantee.
// @termsOfService http://swagger.io/terms/

// @contact.name ClinicEase Support
// @contact.email support@clinicease.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOi..."
func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()
	log.Printf("Starting %s in %s mode...\n", cfg.AppName, cfg.AppEnv)

	// 2. Initialize Database & Auto-Migrations
	db, err := config.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("Fatal: Database migration failed: %v", err)
	}

	// 3. Seed Default Admin User & Sample Clinic
	seedInitialData(db)

	// 4. Initialize JWT Token Manager
	jwtManager := token.NewJWTManager(cfg.JWTSecret, cfg.JWTExpireHours)

	// 5. Initialize Repositories (Data Access Layer)
	userRepo := repository.NewUserRepository(db)
	clinicRepo := repository.NewClinicRepository(db)
	doctorRepo := repository.NewDoctorRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	appointmentRepo := repository.NewAppointmentRepository(db)

	// 6. Initialize Usecases (Business Logic Layer)
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtManager)
	clinicUsecase := usecase.NewClinicUsecase(clinicRepo)
	doctorUsecase := usecase.NewDoctorUsecase(doctorRepo, userRepo, clinicRepo)
	scheduleUsecase := usecase.NewScheduleUsecase(scheduleRepo, doctorRepo, appointmentRepo)
	appointmentUsecase := usecase.NewAppointmentUsecase(appointmentRepo, doctorRepo, scheduleRepo, userRepo)

	// 7. Initialize HTTP Handlers (Delivery Layer)
	authHandler := handler.NewAuthHandler(authUsecase)
	clinicHandler := handler.NewClinicHandler(clinicUsecase)
	doctorHandler := handler.NewDoctorHandler(doctorUsecase)
	scheduleHandler := handler.NewScheduleHandler(scheduleUsecase)
	appointmentHandler := handler.NewAppointmentHandler(appointmentUsecase)

	// 8. Setup Router
	router := deliveryHttp.SetupRouter(deliveryHttp.RouterConfig{
		AuthHandler:        authHandler,
		ClinicHandler:      clinicHandler,
		DoctorHandler:      doctorHandler,
		ScheduleHandler:    scheduleHandler,
		AppointmentHandler: appointmentHandler,
		JWTManager:         jwtManager,
	})

	// 9. Start Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.AppPort),
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on port %s (http://localhost:%s)\n", cfg.AppPort, cfg.AppPort)
		log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html\n", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}

func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")
	return db.AutoMigrate(
		&domain.User{},
		&domain.Clinic{},
		&domain.Doctor{},
		&domain.DoctorSchedule{},
		&domain.Appointment{},
		&domain.MedicalNote{},
	)
}

func seedInitialData(db *gorm.DB) {
	var userCount int64
	db.Model(&domain.User{}).Count(&userCount)
	if userCount == 0 {
		log.Println("Seeding initial admin user and sample data...")

		adminPass, _ := password.Hash("Admin123!")
		adminUser := domain.User{
			Email:        "admin@clinicease.com",
			PasswordHash: adminPass,
			FullName:     "ClinicEase System Administrator",
			Phone:        "+1234567890",
			Role:         domain.RoleAdmin,
		}
		db.Create(&adminUser)

		clinic := domain.Clinic{
			Name:    "Central City Medical Center",
			Address: "123 Healthcare Ave, Suite 400",
			Phone:   "+18005550199",
		}
		db.Create(&clinic)

		// Create a sample doctor user
		docPass, _ := password.Hash("Doctor123!")
		docUser := domain.User{
			Email:        "doctor@clinicease.com",
			PasswordHash: docPass,
			FullName:     "Dr. Stephen Strange",
			Phone:        "+18005550200",
			Role:         domain.RoleDoctor,
		}
		db.Create(&docUser)

		doctor := domain.Doctor{
			UserID:              docUser.ID,
			ClinicID:            clinic.ID,
			Specialization:      "Cardiology",
			SlotDurationMinutes: 30,
			ConsultationFee:     150.00,
			Bio:                 "Senior Cardiologist specializing in preventive heart care and diagnostics.",
		}
		db.Create(&doctor)

		// Create schedules for Monday through Friday: 09:00 to 12:00
		for day := 1; day <= 5; day++ {
			db.Create(&domain.DoctorSchedule{
				DoctorID:  doctor.ID,
				DayOfWeek: day,
				StartTime: "09:00:00",
				EndTime:   "12:00:00",
				IsActive:  true,
			})
		}

		log.Println("Initial data seeded successfully!")
		log.Println("-> Admin account:  admin@clinicease.com  / Admin123!")
		log.Println("-> Doctor account: doctor@clinicease.com / Doctor123!")
	}

	// Always ensure a sample patient exists for testing
	var patientCount int64
	db.Model(&domain.User{}).Where("email IN ?", []string{"patient@example.com", "patient@clinicease.com"}).Count(&patientCount)
	if patientCount == 0 {
		patientPass, _ := password.Hash("Password123!")
		patientUser := domain.User{
			Email:        "patient@example.com",
			PasswordHash: patientPass,
			FullName:     "Jane Doe (Patient)",
			Phone:        "+18005550300",
			Role:         domain.RolePatient,
		}
		if err := db.Create(&patientUser).Error; err == nil {
			log.Println("-> Sample patient seeded: patient@example.com / Password123!")
		}
	}
}
