package main

import (
	"fmt"
	"log"
	"time" // Tambahan: Wajib untuk mengatur durasi MaxAge CORS

	"enterprise-rag-api/config"
	"enterprise-rag-api/models"
	"enterprise-rag-api/routes"

	"github.com/gin-contrib/cors" // Tambahan: Library middleware CORS
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Lakukan Jabat Tangan dengan Database
	config.ConnectDatabase()

	// 2. AutoMigrate: GORM akan otomatis membuatkan tabel AI untuk kita
	fmt.Println("🔄 Memulai migrasi struktur database...")
	err := config.DB.AutoMigrate(&models.Document{}, &models.Task{})
	if err != nil {
		log.Fatal("❌ [FATAL] Gagal membuat tabel: ", err)
	}
	fmt.Println("✅ [SUCCESS] Tabel Database beserta struktur Vector siap digunakan!")

	// 3. Siapkan API Gateway (Router Gin)
	r := gin.Default()

	// ==========================================
	// 🛡️ TAMBAHKAN SATPAM CORS DI SINI
	// ==========================================
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Mengizinkan request dari file HTML lokal
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// ==========================================

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Golang API Gateway is ready for AI Integration!",
		})
	})

	// 4. Mendaftarkan semua route agent (Ini yang memuat rute /task milikmu)
	routes.SetupAgentRoutes(r)

	// 5. Jalankan Server di port 8080
	r.Run(":8080")
}