package config

import (
	"fmt"
	"log"
	"order-service/src/models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Using system environment variables.")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed Connected Posgresql: ", err)
	}

	log.Println("Successfully Connected Posgresql")

	err = db.AutoMigrate(
		&models.Customer{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	)

	if err != nil {
		log.Fatal("Failed to migrate table: ", err)
	}

	log.Println("Successfully migrate table")

	DB = db
}
