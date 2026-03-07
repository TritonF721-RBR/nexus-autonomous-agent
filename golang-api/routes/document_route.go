package routes

import (
	"enterprise-rag-api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupDocumentRoutes mendaftarkan semua rute yang berhubungan dengan dokumen
func SetupDocumentRoutes(r *gin.Engine) {
	// Kita buat grup "/api" agar URL-nya profesional (contoh: /api/documents)
	api := r.Group("/api")
	{
		api.POST("/documents", controllers.IngestDocument)

		api.POST("/search", controllers.SearchDocument)
	}
}
