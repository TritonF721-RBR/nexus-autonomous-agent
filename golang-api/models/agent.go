package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

// Document tetap dipertahankan untuk menyimpan data teks + koordinat vektor RAG
type Document struct {
	ID        uint            `gorm:"primaryKey"`
	Content   string          `gorm:"type:text"`
	Embedding pgvector.Vector `gorm:"type:vector(384)"` // Menyesuaikan library pgvector kamu
}

// Task (BARU) untuk menyimpan riwayat pekerjaan Autonomous Agent ke Database
type Task struct {
	ID          string `gorm:"primaryKey;type:varchar(100)"`
	Instruction string `gorm:"type:text"`
	Status      string `gorm:"type:varchar(50)"` // PROCESSING, COMPLETED, FAILED
	Result      string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
