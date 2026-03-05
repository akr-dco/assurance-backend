package main

import (
	"go-api/config"
	"go-api/middleware"
	"go-api/models"
	"go-api/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Cek environment
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// Load .env sesuai APP_ENV
	err := godotenv.Load(".env." + appEnv)
	if err != nil {
		log.Println("No .env." + appEnv + " file found, fallback to default .env")
		godotenv.Load(".env") // fallback
	}

	config.ConnectDB()
	config.DB.AutoMigrate(

		// ===== MASTER PALING DASAR =====
		&models.MstrCompany{},
		&models.MstrDevice{},
		&models.MstrUser{},
		&models.MstrGroup{},

		// ===== MASTER INSPECTION =====
		&models.MstrInspection{},
		&models.MstrInspectionDetail{},
		&models.MstrInspectionQuestion{},
		&models.MstrInspectionQuestionOption{},
		&models.MstrAnswer{},
		&models.MstrAnswerDetail{},
		&models.MstrTypeTrigger{},
		&models.MstrEventTrigger{},

		// ===== CHAINING =====
		&models.MstrChaining{},
		&models.MstrChainingDetail{},

		// ===== QUESTIONNAIRE =====
		&models.Questionnaire{},
		&models.Question{},
		&models.Option{},
		&models.Answer{},

		// ===== DECLARATION =====
		&models.MstrDeclaration{},

		// ===== TRANSACTION =====
		&models.TrxInspection{},
		&models.TrxInspectionDetail{},
		&models.TrxInspectionAnswer{},

		// ===== TRX ADHOC =====
		&models.TrxSamAdhoc{},

		// ===== SUPPORT =====
		&models.PasswordResetToken{},
	)

	r := gin.Default()

	//Aktifkan middleware CORS sebelum semua route
	r.Use(middleware.CORSMiddleware())
	// ✅ static public dulu
	r.Static("/uploads", "./uploads")
	routes.SetupRoutes(r)

	for _, ri := range r.Routes() {
		println(ri.Method, ri.Path)
	}

	// Jalankan server dengan port dari .env
	r.Run(":" + os.Getenv("SERVER_PORT"))

}
