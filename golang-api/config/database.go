package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB adalah variabel global untuk diakses dari file lain
var DB *gorm.DB

func ConnectDatabase() {
	// Membaca konfigurasi dari file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// Merakit Data Source Name (DSN)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port)

	// Membuka koneksi menggunakan GORM
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ [FATAL] Gagal melakukan koneksi ke Vector Database: ", err)
	}

	// Verifikasi Jabat Tangan (Ping)
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal("❌ [FATAL] Gagal mendapatkan instance database: ", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatal("❌ [FATAL] Database tidak merespon Ping: ", err)
	}

	fmt.Println("✅ [SUCCESS] Berhasil terhubung ke PostgreSQL Vector Database!")
	DB = database
}
