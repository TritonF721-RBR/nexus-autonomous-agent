package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"enterprise-rag-api/config"
	"enterprise-rag-api/models"

	"github.com/gin-gonic/gin"
	"github.com/pgvector/pgvector-go"
)

type DocumentInput struct {
	Content string `json:"content" binding:"required"`
}

type PythonAIResponse struct {
	Dimension int       `json:"dimension"`
	Embedding []float32 `json:"embedding"`
}

// 1. Endpoint: POST /api/ingest
func IngestDocument(c *gin.Context) {
	var input DocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON salah"})
		return
	}

	pythonPayload, _ := json.Marshal(map[string]string{"text": input.Content})
	resp, err := http.Post("http://localhost:8001/api/embed", "application/json", bytes.NewBuffer(pythonPayload))
	if err != nil || resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungi Python"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var aiResult PythonAIResponse
	json.Unmarshal(body, &aiResult)

	doc := models.Document{
		Content:   input.Content,
		Embedding: pgvector.NewVector(aiResult.Embedding),
	}
	config.DB.Create(&doc)

	c.JSON(http.StatusCreated, gin.H{"message": "Data masuk ke Vector DB"})
}

type AgentTaskInput struct {
	Instruction string `json:"instruction" binding:"required"`
}

// 2. Endpoint: POST /task
func AssignTaskToAgent(c *gin.Context) {
	var input AgentTaskInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Instruksi tidak boleh kosong"})
		return
	}

	taskID := fmt.Sprintf("TASK-%d", time.Now().UnixNano())

	// Simpan status AWAL ke Database
	newTask := models.Task{
		ID:          taskID,
		Instruction: input.Instruction,
		Status:      "PROCESSING",
		Result:      "",
	}
	config.DB.Create(&newTask)

	// Jalankan Goroutine Worker di latar belakang
	go backgroundWorker(taskID, input.Instruction)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Tugas diterima. Agen AI sedang memikirkan dan mengeksekusinya di latar belakang.",
		"task_id": taskID,
	})
}

// 3. Endpoint: GET /task/:id/status
func CheckTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	var task models.Task
	if err := config.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task ID tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": task.ID,
		"status":  task.Status,
		"result":  task.Result,
	})
}

// 4. BACKGROUND WORKER GOROUTINE
func backgroundWorker(taskID string, instruction string) {
	pythonPayload, _ := json.Marshal(map[string]string{
		"task_id":     taskID,
		"instruction": instruction,
	})

	resp, err := http.Post("http://localhost:8001/api/agent/execute", "application/json", bytes.NewBuffer(pythonPayload))

	if err != nil || resp.StatusCode != 200 {
		config.DB.Model(&models.Task{}).Where("id = ?", taskID).Updates(models.Task{
			Status: "FAILED",
			Result: "Worker gagal menghubungi Python AI Brain.",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	finalResult := "Tugas selesai, namun format balasan tidak terbaca."
	if res, ok := result["result"].(string); ok {
		finalResult = res
	}

	config.DB.Model(&models.Task{}).Where("id = ?", taskID).Updates(models.Task{
		Status: "COMPLETED",
		Result: finalResult,
	})
}

// ====================================================================
// 5. ENDPOINT INTERNAL (BARU): PENANGKAP TEMBAKAN DARI PYTHON
// ====================================================================
type InternalSearchInput struct {
	Query string `json:"query" binding:"required"`
}

func InternalSearchDocument(c *gin.Context) {
	var input InternalSearchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON salah"})
		return
	}

	// Minta Python mengubah teks query menjadi vektor
	pythonPayload, _ := json.Marshal(map[string]string{"text": input.Query})
	resp, err := http.Post("http://localhost:8001/api/embed", "application/json", bytes.NewBuffer(pythonPayload))
	if err != nil || resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal embed query"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var aiResult PythonAIResponse
	json.Unmarshal(body, &aiResult)

	// Lakukan pencarian Cosine Similarity di PostgreSQL
	var searchResult struct {
		Content string
	}
	queryVector := pgvector.NewVector(aiResult.Embedding)
	config.DB.Raw("SELECT content FROM documents ORDER BY embedding <=> ? LIMIT 1", queryVector).Scan(&searchResult)

	// Kembalikan teks dokumen yang ditemukan ke Agen Python
	c.JSON(http.StatusOK, gin.H{
		"content": searchResult.Content,
	})
}
