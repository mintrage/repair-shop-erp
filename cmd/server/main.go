package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type application struct {
	db *sql.DB
}

type PartInput struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func (app *application) createPartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input PartInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `INSERT INTO parts (name, quantity, price) VALUES ($1, $2, $3) RETURNING id`

	var newID int
	err = app.db.QueryRow(query, input.Name, input.Quantity, input.Price).Scan(&newID)
	if err != nil {
		http.Error(w, "Failed to insert into database", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Created part with ID: %d", newID)
}

func main() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var pingErr error
	for i := 0; i <= 5; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		log.Println("Waiting for database...")
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		log.Fatalf("Database error: %v", pingErr)
	}
	log.Println("Successfully connected to the database!")

	app := application{
		db: db,
	}

	http.HandleFunc("/parts", app.createPartHandler)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
