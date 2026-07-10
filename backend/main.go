package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	"golang.org/x/crypto/bcrypt"
)

// Global database pointer
var db *sql.DB

type UserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Global CORS Middleware to handle browser security and preflight OPTIONS requests cleanly
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func initDatabase() {
	var err error
	// Change "sqlite3" to "sqlite" here as well
	db, err = sql.Open("sqlite", "./users.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	// ... rest of the table creation code stays exactly the same ...

	// Create a SQL table to securely track users if it doesn't exist
	statement, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE,
			password TEXT
		);
	`)
	statement.Exec()
}

func main() {
	initDatabase()
	defer db.Close()

	// Switch Gin to Release Mode when deploying, default is debug mode
	r := gin.Default()
	r.Use(CORSMiddleware())

	// 1. SECURE SIGN UP ENDPOINT
	r.POST("/api/register", func(c *gin.Context) {
		var input UserRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Fields cannot be empty"})
			return
		}

		// Hash password using Bcrypt with a computational cost factor of 10
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Security error"})
			return
		}

		// Insert user securely into SQLite using a prepared SQL query to prevent SQL injections
		query := "INSERT INTO users (username, password) VALUES (?, ?)"
		_, err = db.Exec(query, input.Username, string(hashedPassword))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Username already taken!"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Account created securely!"})
	})

	// 2. SECURE SIGN IN ENDPOINT
	r.POST("/api/login", func(c *gin.Context) {
		var input UserRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Fields cannot be empty"})
			return
		}

		// Find user in the database
		var storedHash string
		query := "SELECT password FROM users WHERE username = ?"
		err := db.QueryRow(query, input.Username).Scan(&storedHash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid username or password"})
			return
		}

		// Verify the incoming password against the hashed string securely
		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid username or password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Login successful! Welcome to the system!"})
	})

	log.Println("Production-grade Auth Engine running on http://localhost:8080")
	r.Run(":8080")
}
