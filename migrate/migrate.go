package main

import (
	"database/sql"
	"log"

	"github.com/joho/godotenv"
	"github.com/umesh/ginapi/config"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	config.ConnectDB()
	defer config.DB.Close()

	// Create tables if they don't exist
	tables := []struct {
		name    string
		creator func(*sql.DB) error
	}{
		{"users", createUsersTable},
		{"products", createProductsTable},
	}

	for _, table := range tables {
		err = table.creator(config.DB)
		if err != nil {
			log.Fatalf("%s table creation failed: %v", table.name, err)
		}
		log.Printf("%s table migration completed successfully", table.name)
	}
}

func createUsersTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`

	_, err := db.Exec(createTableSQL)
	return err
}

func createProductsTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		price DECIMAL(10,2) NOT NULL,
		quantity INT NOT NULL,
		image VARCHAR(255),
		sales_rate DECIMAL(10,2),
		purchase_rate DECIMAL(10,2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`

	_, err := db.Exec(createTableSQL)
	return err
}
