package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

var ErrPokemonNotFound = errors.New("pokemon not found")

// User represents a registered user
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Pokemon represents Pokemon information from API
type Pokemon struct {
	ID      int        `json:"id"`
	Name    string     `json:"name"`
	Height  int        `json:"height"`
	Weight  int        `json:"weight"`
	Sprites Sprites    `json:"sprites"`
	Types   []TypeInfo `json:"types"`
}

type Sprites struct {
	FrontDefault string `json:"front_default"`
}

type TypeInfo struct {
	Type TypeDetail `json:"type"`
}

type TypeDetail struct {
	Name string `json:"name"`
}

// PokemonResponse for API response
type PokemonResponse struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Image  string `json:"image"`
}

// RegistrationRequest for user registration
type RegistrationRequest struct {
	Name string `json:"name" binding:"required"`
}

// PokemonQueryRequest for Pokemon query
type PokemonQueryRequest struct {
	Pokemon string `json:"pokemon" binding:"required"`
}

func init() {
	var err error

	// Build MySQL connection string from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Fallback for local development
	if dbUser == "" {
		dbUser = "root"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbName == "" {
		dbName = "pokebot"
	}

	// MySQL DSN format: user:password@tcp(host:port)/dbname
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to MySQL: %v", err))
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("Failed to ping MySQL: %v", err))
	}

	// Create users table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		panic(fmt.Sprintf("Failed to create table: %v", err))
	}

	fmt.Println("✅ MySQL database connection established")
}

func main() {
	defer db.Close()

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "ok",
			"database": "mysql",
		})
	})

	// User registration endpoint
	router.POST("/api/v1/register", registerUser)

	// Pokemon query endpoint
	router.POST("/api/v1/pokemon", queryPokemon)

	// Get a user by ID
	router.GET("/api/v1/users/:id", getUserByID)

	// Get all users endpoint (for verification)
	router.GET("/api/v1/users", getAllUsers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🤖 PokeBot API Server running on port %s\n", port)
	fmt.Printf("📊 Database: MySQL (%s:%s)\n", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	router.Run(":" + port)
}

// registerUser handles POST /api/v1/register
func registerUser(c *gin.Context) {
	var req RegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Insert user into database
	result, err := db.Exec("INSERT INTO users (name) VALUES (?)", req.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	userID, _ := result.LastInsertId()
	user := User{
		ID:        int(userID),
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	c.JSON(201, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// queryPokemon handles POST /api/v1/pokemon
func queryPokemon(c *gin.Context) {
	var req PokemonQueryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Fetch from Pokemon API
	pokemon, err := fetchPokemonFromAPI(req.Pokemon)
	if errors.Is(err, ErrPokemonNotFound) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "Pokemon API error", "details": err.Error()})
		return
	}

	// Build response
	types := []string{}
	for _, t := range pokemon.Types {
		types = append(types, t.Type.Name)
	}

	typeStr := types[0]
	if len(types) > 1 {
		typeStr = fmt.Sprintf("%s, %s", types[0], types[1])
	}

	response := PokemonResponse{
		Name:   pokemon.Name,
		Type:   typeStr,
		Height: pokemon.Height,
		Weight: pokemon.Weight,
		Image:  pokemon.Sprites.FrontDefault,
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    response,
	})
}

// getAllUsers handles GET /api/v1/users

func getUserByID(c *gin.Context) {
	id := c.Param("id")

	var user User
	err := db.QueryRow("SELECT id, name, created_at FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"name":    user.Name,
		"data":    user,
	})
}

func getAllUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(200, gin.H{
		"total": len(users),
		"users": users,
	})
}

// fetchPokemonFromAPI queries the Pokemon API
func fetchPokemonFromAPI(name string) (*Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Pokemon API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, ErrPokemonNotFound
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var pokemon Pokemon
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &pokemon, nil
}
