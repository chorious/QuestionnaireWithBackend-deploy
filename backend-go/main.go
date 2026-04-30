package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "data.db")
	if err != nil {
		fmt.Println("open db failed:", err)
		os.Exit(1)
	}
	defer db.Close()

	initDB()

	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api")
	{
		api.GET("/version", handleVersion)
		api.POST("/submit", handleSubmit)
		api.GET("/submissions", handleList)
		api.GET("/submissions/export", handleExport)
		api.GET("/stats", handleStats)
	}

	// SPA fallback: serve dist/index.html for non-API routes
	r.NoRoute(func(c *gin.Context) {
		c.File("../dist/index.html")
	})

	port := getenv("PORT", "3000")
	fmt.Println("Server running at http://0.0.0.0:" + port)
	fmt.Println("API endpoints:")
	fmt.Println("  GET  http://localhost:" + port + "/api/version")
	fmt.Println("  POST http://localhost:" + port + "/api/submit")
	fmt.Println("  GET  http://localhost:" + port + "/api/submissions")
	fmt.Println("  GET  http://localhost:" + port + "/api/submissions/export")
	fmt.Println("  GET  http://localhost:" + port + "/api/stats")
	r.Run(":" + port)
}

func initDB() {
	stmt := `
	CREATE TABLE IF NOT EXISTS submissions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		answers TEXT NOT NULL,
		scores TEXT NOT NULL,
		result TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		source TEXT DEFAULT ''
	);
	CREATE INDEX IF NOT EXISTS idx_created_at ON submissions(created_at);
	CREATE INDEX IF NOT EXISTS idx_result ON submissions(result);
	`
	if _, err := db.Exec(stmt); err != nil {
		fmt.Println("init db failed:", err)
		os.Exit(1)
	}
}

func handleVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
}

func handleSubmit(c *gin.Context) {
	var req struct {
		Answers []string       `json:"answers"`
		Scores  map[string]int `json:"scores"`
		Result  string         `json:"result"`
		Source  string         `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	userID := uuid.New().String()
	createdAt := time.Now().UnixMilli()

	_, err := db.Exec(
		"INSERT INTO submissions (id, user_id, answers, scores, result, created_at, source) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, userID, toJSON(req.Answers), toJSON(req.Scores), req.Result, createdAt, req.Source,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[submit] result=%s id=%s time=%d\n", req.Result, id, createdAt)

	c.JSON(http.StatusOK, gin.H{"success": true, "id": id})
}

func handleList(c *gin.Context) {
	rows, err := db.Query("SELECT id, user_id, answers, scores, result, created_at, source FROM submissions ORDER BY created_at DESC LIMIT 10000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var submissions []map[string]any
	for rows.Next() {
		var id, userID, answers, scores, result, source string
		var createdAt int64
		if err := rows.Scan(&id, &userID, &answers, &scores, &result, &createdAt, &source); err != nil {
			continue
		}
		submissions = append(submissions, map[string]any{
			"id":         id,
			"user_id":    userID,
			"answers":    answers,
			"scores":     scores,
			"result":     result,
			"created_at": createdAt,
			"source":     source,
		})
	}

	c.JSON(http.StatusOK, gin.H{"count": len(submissions), "submissions": submissions})
}

func handleExport(c *gin.Context) {
	rows, err := db.Query("SELECT id, user_id, answers, scores, result, created_at, source FROM submissions ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\"submissions.csv\"")

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"id", "user_id", "result", "created_at", "source", "answers", "scores"})

	for rows.Next() {
		var id, userID, answers, scores, result, source string
		var createdAt int64
		if err := rows.Scan(&id, &userID, &answers, &scores, &result, &createdAt, &source); err != nil {
			continue
		}
		writer.Write([]string{id, userID, result, time.UnixMilli(createdAt).Format(time.RFC3339), source, answers, scores})
	}
	writer.Flush()
}

func handleStats(c *gin.Context) {
	var total int
	row := db.QueryRow("SELECT COUNT(*) FROM submissions")
	row.Scan(&total)

	rows, err := db.Query("SELECT result, COUNT(*) as count FROM submissions GROUP BY result ORDER BY count DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var byResult []gin.H
	for rows.Next() {
		var result string
		var count int
		rows.Scan(&result, &count)
		byResult = append(byResult, gin.H{"result": result, "count": count})
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "byResult": byResult})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
