package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mintrage/repair-shop-erp/internal/models"
)

type Application struct {
	DB *sql.DB
}

func (app *Application) AddPartToOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input models.AddPartToOrderInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tx, err := app.DB.Begin()
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

func (app *Application) OrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input models.RepairOrderInput
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
	err = app.DB.QueryRow(query, input.ClientName, input.PhoneNumber, input.DeviceModel, input.ProblemDescription, input.Status).Scan(&newOrderID)
	if err != nil {
		http.Error(w, "Failed to insert into repair_order", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Created order with ID: %d", newOrderID)
}

func (app *Application) PartsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		rows, err := app.DB.Query("SELECT id, name, quantity, price FROM part")
		if err != nil {
			http.Error(w, "Failed to get parts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		parts := []models.Part{}
		for rows.Next() {
			var p models.Part
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
		var input models.Part
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query := `INSERT INTO part (name, quantity, price) VALUES ($1, $2, $3) RETURNING id`
		var newOrderID int
		err = app.DB.QueryRow(query, input.Name, input.Quantity, input.Price).Scan(&newOrderID)
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
