package main

import (
	"fmt"
	"log"

	"enterprise-rag-api/config"
	"enterprise-rag-api/models" // Memanggil folder models
	"enterprise-rag-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Lakukan Jabat Tangan dengan Database
	config.ConnectDatabase()

	// 2. AutoMigrate: GORM akan otomatis membuatkan tabel AI untuk kita
	fmt.Println("🔄 Memulai migrasi struktur database...")
	err := config.DB.AutoMigrate(&models.Document{})
	if err != nil {
		log.Fatal("❌ [FATAL] Gagal membuat tabel Document: ", err)
	}
	fmt.Println("✅ [SUCCESS] Tabel Document beserta struktur Vector siap digunakan!")

	// 3. Siapkan API Gateway (Router Gin)
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Golang API Gateway is ready for AI Integration!",
		})
	})

	// Daftarkan Rute Dokumen ke dalam Gin Engine
	routes.SetupDocumentRoutes(r)

	// 4. Jalankan Server di port 8080
	r.Run(":8080")
}
