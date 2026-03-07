package models

import (
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Document adalah representasi tabel di PostgreSQL
type Document struct {
	gorm.Model
	Content   string          `gorm:"type:text;not null"` // Teks asli yang bisa dibaca manusia
	Embedding pgvector.Vector `gorm:"type:vector(384)"`   // Vektor AI
}
