// package main

// import (
// 	"log"

// 	"github.com/joho/godotenv"
// 	"github.com/umesh/ginapi/config"
// 	"github.com/umesh/ginapi/models"
// )

// func main() {
// 	errenv := godotenv.Load()
// 	if errenv != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// 	config.ConnectDB()

// 	err := config.DB.AutoMigrate(&models.User{})
// 	if err != nil {
// 		log.Fatal("Migration failed: ", err)
// 	}
// 	log.Println("Migration completed successfully")
// }

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

	// Create users table if not exists
	err = createUsersTable(config.DB)
	if err != nil {
		log.Fatal("Table creation failed: ", err)
	}
	log.Println("Table creation completed successfully")
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
