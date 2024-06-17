package crud

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Column struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
	RawName string `json:"rawname"`
}

type Table struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

type CRUDHandler struct {
	DB             *sql.DB
	TableName      string
	ColumnLabels   map[string]string
	Validations    map[string]string
	HiddenColumns  []string
	Table          Table
	linesPerPage   int
	query          string
	RawColumnNames []string
}

func NewCRUDHandler(db *sql.DB) (*CRUDHandler, error) {
	handler := &CRUDHandler{
		DB:             db,
		ColumnLabels:   make(map[string]string),
		Validations:    make(map[string]string),
		HiddenColumns:  []string{},
		linesPerPage:   15,
		query:          "",
		RawColumnNames: []string{},
	}

	return handler, nil
}

func (h *CRUDHandler) Reset() {
	h.TableName = ""
	h.ColumnLabels = make(map[string]string)
	h.Validations = make(map[string]string)
	h.HiddenColumns = []string{}
	h.linesPerPage = 15
	h.query = ""
	h.RawColumnNames = []string{}
}

// Setters to configure our CRUD (each file in the /setup folder will use them)
func (h *CRUDHandler) SetTableName(tableName string)                  { h.TableName = tableName }
func (h *CRUDHandler) SetColumnLabels(columnLabels map[string]string) { h.ColumnLabels = columnLabels }
func (h *CRUDHandler) SetValidations(validations map[string]string)   { h.Validations = validations }
func (h *CRUDHandler) SetLinesPerPage(linesPerPage int)               { h.linesPerPage = linesPerPage }
func (h *CRUDHandler) SetQuery(query string)                          { h.query = query }
func (h *CRUDHandler) SetHiddenColumns(hiddenColumns []string)        { h.HiddenColumns = hiddenColumns }

// we need this function to wait all Setters before starting
func (h *CRUDHandler) Initialize() {
	if h.query != "" {
		log.Println("Query provided, getting table structure from query...")
		h.updateTableStructureFromQuery()
	} else if h.query == "" {
		log.Println("No query provided, getting table structure from database...")
		h.updateTableStructureFromDB()
	} else {
		log.Println("Neither query nor table name provided, couldn´t get table structure.")
	}
}

func (h *CRUDHandler) updateTableStructureFromQuery() {
	h.RawColumnNames = nil
	// Normalize query by removing extra spaces and line breaks
	normalizedQuery := strings.Join(strings.Fields(h.query), " ")

	fromIndex := strings.Index(strings.ToUpper(normalizedQuery), " FROM ")
	if fromIndex == -1 {
		log.Fatal("Failed to find FROM keyword in the query")
	}

	columnSubstring := normalizedQuery[len("SELECT"):fromIndex]
	columnSubstring = strings.TrimSpace(columnSubstring)

	// Remove aliases and extract clean column names
	re := regexp.MustCompile(`(?i)([a-zA-Z0-9_\.]+)(?:\s+AS\s+([a-zA-Z0-9_]+))?`)
	matches := re.FindAllStringSubmatch(columnSubstring, -1)

	var columnNames []string
	var rawColumnNames []string
	for _, match := range matches {
		if len(match) > 1 {
			// Use the alias if it exists, otherwise use the column name
			if len(match) > 2 && match[2] != "" {
				columnNames = append(columnNames, match[2])
			} else {
				// Strip table prefix
				col := match[1]
				if dotIndex := strings.Index(col, "."); dotIndex != -1 {
					col = col[dotIndex+1:]
				}
				columnNames = append(columnNames, col)
			}
		}
		rawColumnNames = append(rawColumnNames, match[1])
	}

	columns := make([]Column, len(columnNames))
	for i, name := range columnNames {
		comment := name
		if label, ok := h.ColumnLabels[name]; ok {
			comment = label
		}
		columns[i] = Column{Name: name, Comment: comment}
	}

	h.Table = Table{Name: h.TableName, Columns: columns}
	h.RawColumnNames = rawColumnNames

	//log.Println("matches: ", matches)
	//log.Println("h.Table: ", h.Table)
	log.Println("columnNames: ", columnNames)
	log.Println("rawColumnNames: ", rawColumnNames)
	fmt.Println("====================================================")
}

func (h *CRUDHandler) updateTableStructureFromDB() {
	query := fmt.Sprintf("SHOW FULL COLUMNS FROM %s", h.TableName)
	rows, err := h.DB.Query(query)
	h.RawColumnNames = nil
	if err != nil {
		log.Printf("Error querying table structure: %s", err)
		return
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var field, colType, collation, null, key, extra, privileges, comment sql.NullString
		if err := rows.Scan(&field, &colType, &collation, &null, &key, &sql.RawBytes{}, &extra, &privileges, &comment); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		columns = append(columns, Column{
			Name:    field.String,
			Type:    colType.String,
			Comment: toTitleCase(field.String), // Use the actual tranformed column name as the Comment
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
	}

	h.Table = Table{
		Name:    h.TableName,
		Columns: columns,
	}

	log.Println("in CRUD.GO executing updateTableStructureFromDB() - \n Extracted column names from DB:", columns)
	fmt.Println("====================================================")
}

/* HELPER to transform the snake case to Title Case */
func toTitleCase(str string) string {
	words := strings.Split(str, "_")
	for i := 0; i < len(words); i++ {
		words[i] = cases.Title(language.English).String(words[i])
	}
	return strings.Join(words, " ")
}

func (h *CRUDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.addRecordHandler(w, r)
	default:
		h.recordsHandler(w, r)
	}
}

// handle the extraction of parameters from the request
func (h *CRUDHandler) extractParams(r *http.Request) (int, string, string, string, string) {
	params := r.URL.Query()

	pageStr := params.Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		page = 0
	}

	search := params.Get("search")
	id := params.Get("id")
	sortColumn := params.Get("sort")
	sortOrder := params.Get("order")

	return page, search, id, sortColumn, sortOrder
}

// build the SQL query
func (h *CRUDHandler) buildQuery(page int, search string, id string, sortColumn string, sortOrder string) string {
	h.Initialize()
	where := ""
	if search != "" {
		where = " WHERE " + h.buildSearchQuery(search, h.RawColumnNames)
	}
	if id != "" {
		where = " WHERE " + h.TableName + ".id = " + id
	}

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

	fmt.Println("Query :", query)
	fmt.Println("====================================================")
	return query
}

// fetch the records from the database
func (h *CRUDHandler) fetchRecords(query string) ([]map[string]interface{}, []string, error) {
	rows, err := h.DB.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var records []map[string]interface{}
	for rows.Next() {
		record := make(map[string]interface{})
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(sql.NullString)
		}
		if err := rows.Scan(values...); err != nil {
			return nil, nil, err
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

	return records, columns, nil
}

// calculate the total pages
func (h *CRUDHandler) calculateTotalPages(id string, limit int) int {
	totalPages := 0
	if id == "" {
		totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", h.TableName)
		var totalRecords int
		h.DB.QueryRow(totalQuery).Scan(&totalRecords)
		totalPages = (totalRecords + limit - 1) / limit
	}
	return totalPages
}

// main handler function
func (h *CRUDHandler) recordsHandler(w http.ResponseWriter, r *http.Request) {
	page, search, id, sortColumn, sortOrder := h.extractParams(r)
	query := h.buildQuery(page, search, id, sortColumn, sortOrder)
	records, columns, err := h.fetchRecords(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	totalPages := h.calculateTotalPages(id, h.linesPerPage)

	response := map[string]interface{}{
		"records":     filterHiddenColumns(records, h.HiddenColumns),
		"columns":     filterHiddenColumnsInColumns(columns, h.HiddenColumns),
		"currentPage": page,
		"totalPages":  totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// builds and returns a string with the search string being searched in all columns of the Table
func (h *CRUDHandler) buildSearchQuery(search string, rawColumnNames []string) string {
	conditions := []string{}
	ret := ""
	for _, columnName := range rawColumnNames {
		if columnName == "id" { // to avoid ambiguity in the select
			columnName = h.TableName + "." + "id"
		}
		condition := fmt.Sprintf("%s LIKE '%%%s%%'", columnName, search)
		conditions = append(conditions, condition)
	}
	ret = strings.Join(conditions, " OR ")

	fmt.Println("Search condition = ", ret)

	return ret
}

func filterHiddenColumns(records []map[string]interface{}, hiddenColumns []string) []map[string]interface{} {
	filteredRecords := []map[string]interface{}{}
	for _, record := range records {
		filteredRecord := map[string]interface{}{}
		for key, value := range record {
			if !contains(hiddenColumns, key) {
				filteredRecord[key] = value
			}
			filteredRecord["id"] = record["id"] // Ensure ID is always included
		}
		filteredRecords = append(filteredRecords, filteredRecord)
	}
	return filteredRecords
}

func filterHiddenColumnsInColumns(columns []string, hiddenColumns []string) []string {
	filteredColumns := []string{}
	for _, column := range columns {
		if !contains(hiddenColumns, column) {
			filteredColumns = append(filteredColumns, column)
		}
	}
	return filteredColumns
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (h *CRUDHandler) addRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert record to SQL insert statement and execute
	columns, values := getSQLInsertColumnsAndValues(record)
	query := fmt.Sprintf("INSERT INTO user_details (%s) VALUES (%s)", columns, values)
	_, err = h.DB.Exec(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

func (h *CRUDHandler) updateRecordHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the ID from the form data
	id := r.FormValue("id")

	// Convert form data to a map
	record := make(map[string]interface{})
	for key, values := range r.Form {
		// If the form field has multiple values (e.g., multiple select),
		// we store all of them, otherwise just store the single value
		if len(values) > 1 {
			record[key] = values
		} else {
			record[key] = values[0]
		}
	}

	// Convert record to SQL update statement and execute
	setClause := getSQLUpdateSetClause(record)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = %s", h.TableName, setClause, id)
	_, err := h.DB.Exec(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)
}

func getSQLInsertColumnsAndValues(record map[string]interface{}) (string, string) {
	var columns, values []string
	for col, val := range record {
		columns = append(columns, col)
		values = append(values, fmt.Sprintf("'%v'", val))
	}
	return strings.Join(columns, ", "), strings.Join(values, ", ")
}

func getSQLUpdateSetClause(record map[string]interface{}) string {
	var setClause []string
	for col, val := range record {
		setClause = append(setClause, fmt.Sprintf("%s = '%v'", col, val))
	}
	return strings.Join(setClause, ", ")
}

/* HELPER Function for Debug purposes */
func showParams(p map[string][]string) {
	for key, values := range p {
		fmt.Printf("Key: %s ", key)
		for _, value := range values {
			fmt.Printf("Value: %s\n", value)
		}
	}
}
