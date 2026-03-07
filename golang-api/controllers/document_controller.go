package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"enterprise-rag-api/config"
	"enterprise-rag-api/models"

	"github.com/gin-gonic/gin"
	"github.com/pgvector/pgvector-go"
)

// Struktur dari User ke Golang
type DocumentInput struct {
	Content string `json:"content" binding:"required"`
}

// Struktur untuk menangkap balasan dari Python AI
type PythonAIResponse struct {
	Dimension int       `json:"dimension"`
	Embedding []float32 `json:"embedding"` // Harus float32 untuk masuk ke pgvector
}

func IngestDocument(c *gin.Context) {
	var input DocumentInput

	// 1. Validasi Input JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON salah atau 'content' kosong"})
		return
	}

	// 2. Siapkan paket data untuk dikirim ke Python
	pythonPayload, _ := json.Marshal(map[string]string{
		"text": input.Content,
	})

	// 3. Golang "Menelepon" Microservice Python di port 8001
	resp, err := http.Post("http://localhost:8001/api/embed", "application/json", bytes.NewBuffer(pythonPayload))
	if err != nil || resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungi Mesin AI Python"})
		return
	}
	defer resp.Body.Close()

	// 4. Baca balasan Vektor (384 angka) dari Python
	body, _ := io.ReadAll(resp.Body)
	var aiResult PythonAIResponse
	json.Unmarshal(body, &aiResult)

	// 5. Rakit Dokumen lengkap (Teks + Vektor AI)
	doc := models.Document{
		Content: input.Content,
		// Kita ubah array float32 menjadi format Vector yang dipahami PostgreSQL
		Embedding: pgvector.NewVector(aiResult.Embedding),
	}

	// 6. Simpan ke Database (Perhatikan: Fungsi .Omit() sudah kita buang karena sekarang vektornya ADA isinya!)
	if err := config.DB.Create(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan vektor ke database"})
		return
	}

	// 7. Berikan respon sukses ke User
	c.JSON(http.StatusCreated, gin.H{
		"message": "Dokumen berhasil diproses oleh AI dan disimpan ke Vector Database!",
		"data": gin.H{
			"id":        doc.ID,
			"content":   doc.Content,
			"dimension": aiResult.Dimension, // Menampilkan angka 384 sebagai bukti
		},
	})

}

type SearchInput struct {
	Query string `json:"query" binding:"required"`
}

// Struktur untuk membaca balasan dari Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func SearchDocument(c *gin.Context) {
	var input SearchInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON salah"})
		return
	}

	// 1. Minta Python mengubah 'Pertanyaan' menjadi Vektor 384 dimensi
	pythonPayload, _ := json.Marshal(map[string]string{"text": input.Query})
	resp, err := http.Post("http://localhost:8001/api/embed", "application/json", bytes.NewBuffer(pythonPayload))
	if err != nil || resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungi Python AI"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var aiResult PythonAIResponse
	json.Unmarshal(body, &aiResult)

	// 2. Lakukan Pencarian Cosine Distance (<=>) di PostgreSQL
	var searchResult struct {
		Content string
	}
	queryVector := pgvector.NewVector(aiResult.Embedding)
	config.DB.Raw("SELECT content FROM documents ORDER BY embedding <=> ? LIMIT 1", queryVector).Scan(&searchResult)

	// ==========================================
	// 3. FASE GENERATION (MEMANGGIL GEMINI AI)
	// ==========================================

	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))

	// Merakit Instruksi (Prompt Engineering)
	promptText := fmt.Sprintf(`Anda adalah asisten AI Perusahaan yang sangat cerdas. 
Tugas Anda adalah menjawab pertanyaan user HANYA berdasarkan konteks dokumen yang diberikan. 
Jika dokumen tidak mengandung jawabannya, katakan "Saya tidak menemukan informasi tersebut di dalam dokumen."

Konteks Dokumen: %s

Pertanyaan User: %s

Jawaban Anda:`, searchResult.Content, input.Query)

	// Merakit JSON Payload sesuai standar Google Gemini
	geminiPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": promptText},
				},
			},
		},
	}
	geminiBody, _ := json.Marshal(geminiPayload)

	// Menembakkan Request ke Server Google
	geminiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	geminiResp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(geminiBody))

	if err != nil || geminiResp.StatusCode != 200 {
		// ALAT SADAP SUPER: Membaca surat penolakan dari Google
		fmt.Println("🚨 STATUS DARI GOOGLE:", geminiResp.Status)
		if geminiResp != nil {
			errorBody, _ := io.ReadAll(geminiResp.Body)
			fmt.Println("🚨 ALASAN PENOLAKAN:", string(errorBody))
		} else {
			fmt.Println("🚨 ERROR GOLANG:", err)
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungi Gemini AI"})
		return
	}
	defer geminiResp.Body.Close()

	// Membaca dan Mengekstrak jawaban Gemini
	geminiRespBody, _ := io.ReadAll(geminiResp.Body)
	var geminiResult GeminiResponse
	json.Unmarshal(geminiRespBody, &geminiResult)

	// Ekstrak teks final dari struktur JSON
	finalAnswer := "Maaf, AI gagal merangkai kalimat."
	if len(geminiResult.Candidates) > 0 && len(geminiResult.Candidates[0].Content.Parts) > 0 {
		finalAnswer = geminiResult.Candidates[0].Content.Parts[0].Text
	}

	// 4. Kembalikan hasil sempurna ke User
	c.JSON(http.StatusOK, gin.H{
		"message": "Operasi RAG Selesai!",
		"data": gin.H{
			"question":       input.Query,
			"retrieved_docs": searchResult.Content,
			"ai_answer":      finalAnswer, // Ini adalah suara dari Gemini!
		},
	})
}
