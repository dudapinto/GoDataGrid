package server

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"datagrid/crud"
	"datagrid/server/loaders"
)

func SetupRoutes(db *sql.DB) {

	handler, err := crud.NewCRUDHandler(db)
	if err != nil {
		log.Fatal(err)
	}

	// Route for the menu
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/menu.html")
	})

	// Route for CRUD operations
	http.HandleFunc("/crud", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.URL.Query().Get("table")
		if tableName == "" {
			http.Error(w, "Table name is required", http.StatusBadRequest)
			return
		}

		// Load the handler setup dynamically
		setupFunc := loaders.LoadHandlerSetupFunc(tableName)
		if setupFunc == nil {
			http.Error(w, "Unknown table", http.StatusBadRequest)
			return
		}

		err = setupFunc(handler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		handler.Initialize()

		/* 		log.Println("in server.go - After Initialize(): handler = ", handler)
		   		fmt.Println("====================================================")
		   		log.Println("handler.TableName = ", handler.TableName)
		   		log.Println("handler.Table.Columns = ", handler.Table.Columns)
		   		log.Println("Query = ", handler.GetQuery())
		   		fmt.Println("====================================================") */
		// Serve the HTML template with the handler data
		tmpl, err := template.ParseFiles("crud/templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			TableName string
			Columns   []crud.Column
		}{
			TableName: handler.TableName,
			Columns:   handler.Table.Columns,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Serve API requests

	http.Handle("/api/records", handler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}
