package crud

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Column struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
}

type Table struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

type CRUDHandler struct {
	DB           *sql.DB
	TableName    string
	ColumnLabels map[string]string
	Validations  map[string]string
	Table        Table
	linesPerPage int
	query        string
}

func NewCRUDHandler(db *sql.DB) (*CRUDHandler, error) {
	handler := &CRUDHandler{DB: db}
	return handler, nil
}

func (h *CRUDHandler) SetTableName(tableName string) { h.TableName = tableName }

func (h *CRUDHandler) SetColumnLabels(columnLabels map[string]string) { h.ColumnLabels = columnLabels }

func (h *CRUDHandler) SetValidations(validations map[string]string) { h.Validations = validations }

func (h *CRUDHandler) SetLinesPerPage(linesPerPage int) { h.linesPerPage = linesPerPage }

func (h *CRUDHandler) SetQuery(query string) { h.query = query }

func (h *CRUDHandler) Initialize() {
	if h.query != "" {
		log.Println("in CRUD.GO executing Initialise() \n Query provided, updating table structure from query...\n", h.query)
		fmt.Println("======================================================")
		h.updateTableStructureFromQuery()
	} else if h.TableName != "" {
		log.Println("in CRUD.GO executing Initialise() \n No query provided, updating table structure from database...")
		fmt.Println("============================================================")
		h.updateTableStructureFromDB()
	} else {
		log.Println("in CRUD.GO executing Initialise() \n Neither query nor table name provided, cannot update table structure.")
		fmt.Println("=====================================================================")
	}
}

func (h *CRUDHandler) updateTableStructureFromQuery() {

	fromIndex := strings.Index(strings.ToUpper(h.query), " FROM ")
	if fromIndex == -1 {
		log.Fatal("Failed to find FROM keyword in the query")
	}

	columnSubstring := h.query[len("SELECT"):fromIndex]

	columnNames := strings.Split(columnSubstring, ",")
	for i, col := range columnNames {
		columnNames[i] = strings.TrimSpace(col)
	}

	log.Println("Extracted column names:", columnNames)

	columns := make([]Column, len(columnNames))
	for i, col := range columnNames {
		columns[i] = Column{
			Name:    col,
			Comment: h.ColumnLabels[col],
		}
	}

	h.Table = Table{
		Name:    h.TableName,
		Columns: columns,
	}

	log.Println("With labels: ", h.Table)
	fmt.Println("====================================================")
}

func (h *CRUDHandler) updateTableStructureFromDB() {
	query := fmt.Sprintf("SHOW FULL COLUMNS FROM %s", h.TableName)
	rows, err := h.DB.Query(query)

	if err != nil {
		log.Printf("Error querying table structure: %s", err)
		return
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var field, colType, collation, null, key, extra, privileges, comment sql.NullString
		if err := rows.Scan(&field, &colType, &collation, &null, &key, &extra, &privileges, &comment); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		columns = append(columns, Column{
			Name:    field.String,
			Type:    colType.String,
			Comment: comment.String,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
	}

	h.Table = Table{
		Name:    h.TableName,
		Columns: columns,
	}

	log.Println("STEP 1 - No query provided")
	fmt.Println("====================================================")
	log.Println("in CRUD.GO executing updateTableStructureFromDB() - Extracted column names from DB:", columns)
	fmt.Println("====================================================")
}

func (h *CRUDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.addRecordHandler(w, r)
	default:
		h.recordsHandler(w, r)
	}
}

func (h *CRUDHandler) AddRecord(w http.ResponseWriter, r *http.Request) {
	h.addRecordHandler(w, r)
}

func (h *CRUDHandler) GetQuery() string {
	return h.query
}

func (h *CRUDHandler) recordsHandler(w http.ResponseWriter, r *http.Request) {
	// Get all query parameters
	params := r.URL.Query()

	// Iterate over the query parameters and print them
	for key, values := range params {
		// Print the key and all its values
		fmt.Printf("Key: %s ", key)
		for _, value := range values {
			fmt.Printf("Value: %s\n", value)
		}
	}

	pageStr := params.Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		page = 0
	}

	search := params.Get("search")
	where := ""
	if search != "" {
		where = " WHERE " + h.buildSearchQuery(search)
	}

	id := params.Get("id")
	if id != "" {
		where = " WHERE id = " + id
	}

	sortColumn := params.Get("sort")
	sortOrder := params.Get("order")
	orderBy := ""
	if sortColumn != "" && (sortOrder == "asc" || sortOrder == "desc") {
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)
	}

	limit := h.linesPerPage
	offset := page * limit

	query := h.query

	if query == "" {
		query = fmt.Sprintf("SELECT * FROM %s %s %s LIMIT %d OFFSET %d", h.TableName, where, orderBy, limit, offset)
	} else {
		query = fmt.Sprintf("%s %s %s LIMIT %d OFFSET %d", query, where, orderBy, limit, offset)
	}

	//fmt.Println("sortColumn:", sortColumn)
	//fmt.Println("sortOrder:", sortOrder)

	rows, err := h.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var records []map[string]interface{}
	for rows.Next() {
		record := make(map[string]interface{})
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(sql.NullString)
		}
		if err := rows.Scan(values...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i, column := range columns {
			if value, ok := values[i].(*sql.NullString); ok && value.Valid {
				record[column] = value.String
			} else {
				record[column] = nil
			}
		}
		records = append(records, record)
	}

	totalPages := 0
	if id == "" {
		totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", h.TableName)
		var totalRecords int
		h.DB.QueryRow(totalQuery).Scan(&totalRecords)
		totalPages = (totalRecords + limit - 1) / limit
	}

	response := map[string]interface{}{
		"records":     records,
		"currentPage": page,
		"totalPages":  totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CRUDHandler) addRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	columns := make([]string, 0, len(record))
	values := make([]interface{}, 0, len(record))
	placeholders := make([]string, 0, len(record))

	for col, val := range record {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", h.TableName, strings.Join(columns, ","), strings.Join(placeholders, ","))
	_, err := h.DB.Exec(query, values...)
	if err != nil {
		http.Error(w, "Failed to add record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *CRUDHandler) buildSearchQuery(search string) string {
	var conditions []string
	for _, column := range h.Table.Columns {
		condition := fmt.Sprintf("%s LIKE '%%%s%%'", column.Name, search)
		conditions = append(conditions, condition)
	}
	return strings.Join(conditions, " OR ")
}
