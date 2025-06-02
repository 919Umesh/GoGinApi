package main

import (
	"database/sql"
	"log"

	"github.com/joho/godotenv"
	"github.com/umesh/ginapi/config"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()
	defer config.DB.Close()

	tables := []struct {
		name    string
		creator func(*sql.DB) error
	}{
		{"users", createUsersTable},
		{"products", createProductsTable},
		{"venues", createVenuesTable},
	}

	for _, table := range tables {
		err = table.creator(config.DB)
		if err != nil {
			log.Fatalf("%s table creation failed: %v", table.name, err)
		}
		log.Printf("%s table migration completed successfully", table.name)
	}
}

// Create the User Table
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

// Create the product table
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

	if _, err := db.Exec(createTableSQL); err != nil {
		return err
	}
	return nil
}

// Create the venue table
func createVenuesTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS venues (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		location VARCHAR(255) NOT NULL,
		size VARCHAR(100) NOT NULL,
		image VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(createTableSQL)
	return err
}

func createOrderTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
)`

	_, err := db.Exec(createTableSQL)
	return err
}

func createOrderItemsTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
)`

	_, err := db.Exec(createTableSQL)
	return err
}
