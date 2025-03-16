package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Points   int    `json:"points"`
}

// Girl represents a girl to bet on
type Girl struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Bet represents a bet placed by a user
type Bet struct {
	ID        int `json:"id"`
	UserID    int `json:"user_id"`
	GirlID    int `json:"girl_id"`
	BetAmount int `json:"bet_amount"`
}

func main() {
	// Connect to PostgreSQL
	connStr := "user=postgres dbname=betting sslmode=disable password=yourpassword"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize database tables
	initDB()

	// Set up routes
	r := mux.NewRouter()
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/place-bet", placeBetHandler).Methods("POST")
	r.HandleFunc("/girls", getGirlsHandler).Methods("GET")

	// Serve static files (HTML, CSS, JS)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	// Start server
	fmt.Println("Server running on :3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

// Initialize database tables
func initDB() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			points INTEGER DEFAULT 100
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS girls (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			girl_id INTEGER REFERENCES girls(id),
			bet_amount INTEGER NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert sample girls if the table is empty
	var count int
	db.QueryRow("SELECT COUNT(*) FROM girls").Scan(&count)
	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO girls (name) VALUES
			('Alice'),
			('Beth'),
			('Cathy')
		`)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Register a new user
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Insert user into database
	query := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	err = db.QueryRow(query, user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Place a bet
func placeBetHandler(w http.ResponseWriter, r *http.Request) {
	var bet Bet
	err := json.NewDecoder(r.Body).Decode(&bet)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Insert bet into database
	query := `INSERT INTO bets (user_id, girl_id, bet_amount) VALUES ($1, $2, $3)`
	_, err = db.Exec(query, bet.UserID, bet.GirlID, bet.BetAmount)
	if err != nil {
		http.Error(w, "Failed to place bet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Get all girls
func getGirlsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM girls")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var girls []Girl
	for rows.Next() {
		var girl Girl
		err := rows.Scan(&girl.ID, &girl.Name)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		girls = append(girls, girl)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(girls)
}
