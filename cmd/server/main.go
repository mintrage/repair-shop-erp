package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mintrage/repair-shop-erp/internal/handlers"
)

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

	app := handlers.Application{
		DB: db,
	}

	http.HandleFunc("/parts", app.PartsHandler)
	http.HandleFunc("/orders", app.OrdersHandler)
	http.HandleFunc("/order-parts", app.AddPartToOrderHandler)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
