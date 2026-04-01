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

type RepairOrder struct {
	ID                 int    `json:"id"`
	ClientName         string `json:"client_name"`
	PhoneNumber        string `json:"phone_number"`
	DeviceModel        string `json:"device_model"`
	ProblemDescription string `json:"problem_description"`
	Status             string `json:"status"`
}

type RepairOrderPartInput struct {
	PartID   int `json:"part_id"`
	Quantity int `json:"quantity"`
}

type RepairOrderInput struct {
	ClientName         string                 `json:"client_name"`
	PhoneNumber        string                 `json:"phone_number"`
	DeviceModel        string                 `json:"device_model"`
	ProblemDescription string                 `json:"problem_description"`
	Status             string                 `json:"status"`
	Parts              []RepairOrderPartInput `json:"items"`
}

type AddPartToOrderInput struct {
	OrderID  int `json:"order_id"`
	PartID   int `json:"part_id"`
	Quantity int `json:"quantity"`
}

func (app *application) addPartToOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input AddPartToOrderInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tx, err := app.db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	defer tx.Rollback()
	insertQuery := `INSERT INTO repair_order_part (order_id, part_id, quantity) VALUES ($1, $2, $3)`
	_, err = tx.Exec(insertQuery, input.OrderID, input.PartID, input.Quantity)
	if err != nil {
		http.Error(w, "Failed to insert into repair_order_part", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	updateQuery := `UPDATE part SET quantity = quantity - $1 WHERE id = $2`
	_, err = tx.Exec(updateQuery, input.Quantity, input.PartID)
	if err != nil {
		http.Error(w, "Failed to update repair_order_part", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	tx.Commit()
	w.WriteHeader(http.StatusOK)
}

func (app *application) ordersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input RepairOrderInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := `
	INSERT INTO repair_order (client_name, phone_number, device_model, problem_description, status) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id
	`
	var newOrderID int
	if input.Status == "" {
		input.Status = "Принят"
	}
	err = app.db.QueryRow(query, input.ClientName, input.PhoneNumber, input.DeviceModel, input.ProblemDescription, input.Status).Scan(&newOrderID)
	if err != nil {
		http.Error(w, "Failed to insert into repair_order", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Created order with ID: %d", newOrderID)
}

func (app *application) partsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		rows, err := app.db.Query("SELECT id, name, quantity, price FROM part")
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
		query := `INSERT INTO part (name, quantity, price) VALUES ($1, $2, $3) RETURNING id`
		var newOrderID int
		err = app.db.QueryRow(query, input.Name, input.Quantity, input.Price).Scan(&newOrderID)
		if err != nil {
			http.Error(w, "Failed to insert into database", http.StatusInternalServerError)
			log.Printf("DB error: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Created part with ID: %d", newOrderID)

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
	http.HandleFunc("/order-parts", app.addPartToOrderHandler)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
