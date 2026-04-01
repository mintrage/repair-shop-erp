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

type Part struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type Order struct {
	ID                 int       `json:"id"`
	ClientName         string    `json:"client_name"`
	PhoneNumber        string    `json:"phone_number"`
	DeviceModel        string    `json:"device_model"`
	ProblemDescription string    `json:"problem_description"`
	Status             string    `json:"status"`
	CreationDate       time.Time `json:"creation_date"`
}

func (app *application) ordersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input Order
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `
	INSERT INTO orders (client_name, phone_number, device_model, problem_description, status) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id
	`
	var newID int
	if input.Status == "" {
		input.Status = "Принят"
	}
	err = app.db.QueryRow(query, input.ClientName, input.PhoneNumber, input.DeviceModel, input.ProblemDescription, input.Status).Scan(&newID)
	if err != nil {
		http.Error(w, "Failed to insert into database", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Created order with ID: %d", newID)
}

func (app *application) partsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		rows, err := app.db.Query("SELECT id, name, quantity, price FROM parts")
		if err != nil {
			http.Error(w, "Failed to get parts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		parts := []Part{}
		for rows.Next() {
			var p Part
			if err := rows.Scan(&p.ID, &p.Name, &p.Quantity, &p.Price); err != nil {
				http.Error(w, "Failed to parse part", http.StatusInternalServerError)
				return
			}
			parts = append(parts, p)
		}
		if err = rows.Err(); err != nil {
			http.Error(w, "Failed to parse parts", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parts)

	case http.MethodPost:
		var input Part
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

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	http.HandleFunc("/parts", app.partsHandler)
	http.HandleFunc("/orders", app.ordersHandler)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
