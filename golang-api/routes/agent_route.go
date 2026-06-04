package routes

import (
	"enterprise-rag-api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupAgentRoutes bertugas mendaftarkan semua endpoint yang berkaitan dengan AI Agent
func SetupAgentRoutes(router *gin.Engine) {

	// Endpoint untuk Ingest Dokumen (Memasukkan data ke Vector DB)
	router.POST("/api/ingest", controllers.IngestDocument)

	// Endpoint untuk memberikan tugas baru ke AI Agent
	router.POST("/task", controllers.AssignTaskToAgent)

	// Endpoint untuk mengecek status dan melihat hasil kerja AI Agent
	router.GET("/task/:id/status", controllers.CheckTaskStatus)

	// ====================================================================
	// ENDPOINT INTERNAL (BARU): Rute yang diakses diam-diam oleh Python
	// ====================================================================
	router.POST("/api/internal/search", controllers.InternalSearchDocument)
}
