package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"datagrid/server"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Database connection
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set up routes and start the server
	server.SetupRoutes(db)

	serverType := runtime.GOOS
	clearCmd := "clear"
	if serverType == "windows" {
		clearCmd = "cls"
	}

	cmd := exec.Command(clearCmd)
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Println("This is a", serverType, "server.")

	// Start HTTP server
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
