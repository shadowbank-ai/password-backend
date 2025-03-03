package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Database connection
var db *sql.DB

func initDB() {
	// Fetch DATABASE_URL from environment variables
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ ERROR: DATABASE_URL is not set!")
	}

	fmt.Println("🔍 Trying to connect to:", dsn) // Debugging

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Database connection error:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("❌ Database unreachable:", err)
	}

	fmt.Println("✅ Connected to database!")
}

// Encrypt password before storing
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Store password
func savePassword(c *gin.Context) {
	var input struct {
		Site     string `json:"site"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
		return
	}

	_, err = db.Exec("INSERT INTO passwords (site, username, password) VALUES ($1, $2, $3)", input.Site, input.Username, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password saved successfully"})
}

// Fetch all passwords
func getPasswords(c *gin.Context) {
	rows, err := db.Query("SELECT site, username, password FROM passwords")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch passwords"})
		return
	}
	defer rows.Close()

	var passwords []map[string]string
	for rows.Next() {
		var site, username, password string
		if err := rows.Scan(&site, &username, &password); err != nil {
			continue
		}
		passwords = append(passwords, map[string]string{"site": site, "username": username, "password": password})
	}

	c.JSON(http.StatusOK, passwords)
}

// Main function
func main() {
	initDB()

	r := gin.Default()

	r.POST("/save", savePassword)
	r.GET("/passwords", getPasswords)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
