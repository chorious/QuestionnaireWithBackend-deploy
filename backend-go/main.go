package main

import (
	"database/sql"
	"embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

//go:embed dist
var distFS embed.FS

var db *sql.DB
var adminToken string

func init() {
	adminToken = os.Getenv("ADMIN_TOKEN")
}

func main() {
	var err error
	dbPath := getenv("DB_PATH", "./data/data.db")
	if err := os.MkdirAll("./data", 0755); err != nil {
		fmt.Println("create data dir failed:", err)
		os.Exit(1)
	}
	db, err = sql.Open("sqlite", dbPath)
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
	}

	admin := api.Group("", adminAuth())
	{
		admin.GET("/submissions", handleList)
		admin.GET("/submissions/export", handleExport)
		admin.GET("/stats", handleStats)
	}

	// Serve embedded static files
	staticFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		fmt.Println("static fs failed:", err)
		os.Exit(1)
	}
	r.StaticFS("/", http.FS(staticFS))

	// SPA fallback
	r.NoRoute(func(c *gin.Context) {
		data, err := distFS.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "index.html not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
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
	// Migrate: add name and phone columns if not exist (backward compatible)
	_, _ = db.Exec("ALTER TABLE submissions ADD COLUMN name TEXT DEFAULT ''")
	_, _ = db.Exec("ALTER TABLE submissions ADD COLUMN phone TEXT DEFAULT ''")
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
		UserID  string         `json:"user_id"`
		Name    string         `json:"name"`
		Phone   string         `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	phoneRe := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRe.MatchString(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number"})
		return
	}

	id := uuid.New().String()
	userID := req.UserID
	if userID == "" {
		userID = uuid.New().String()
	}
	createdAt := time.Now().UnixMilli()

	_, err := db.Exec(
		"INSERT INTO submissions (id, user_id, answers, scores, result, created_at, source, name, phone) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, userID, toJSON(req.Answers), toJSON(req.Scores), req.Result, createdAt, req.Source, req.Name, req.Phone,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[submit] result=%s user_id=%s name=%s id=%s time=%d\n", req.Result, userID, req.Name, id, createdAt)

	c.JSON(http.StatusOK, gin.H{"success": true, "id": id, "user_id": userID})
}

func handleList(c *gin.Context) {
	rows, err := db.Query("SELECT id, user_id, answers, scores, result, created_at, source, name, phone FROM submissions ORDER BY created_at DESC LIMIT 10000")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var submissions []map[string]any
	for rows.Next() {
		var id, userID, answers, scores, result, source, name, phone string
		var createdAt int64
		if err := rows.Scan(&id, &userID, &answers, &scores, &result, &createdAt, &source, &name, &phone); err != nil {
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
			"name":       name,
			"phone":      phone,
		})
	}

	c.JSON(http.StatusOK, gin.H{"count": len(submissions), "submissions": submissions})
}

func handleExport(c *gin.Context) {
	rows, err := db.Query("SELECT id, user_id, answers, scores, result, created_at, source, name, phone FROM submissions ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\"submissions.csv\"")

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"id", "user_id", "name", "phone", "result", "created_at", "source", "answers", "scores"})

	for rows.Next() {
		var id, userID, answers, scores, result, source, name, phone string
		var createdAt int64
		if err := rows.Scan(&id, &userID, &answers, &scores, &result, &createdAt, &source, &name, &phone); err != nil {
			continue
		}
		writer.Write([]string{id, userID, name, phone, result, time.UnixMilli(createdAt).Format(time.RFC3339), source, answers, scores})
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

func adminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if adminToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "admin token not configured"})
			return
		}
		if c.GetHeader("X-Admin-Token") != adminToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Token")
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
