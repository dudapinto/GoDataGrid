package setup

import "datagrid/crud"

func UserDetailsSetup(handler *crud.CRUDHandler) error {

	query := `SELECT
	    user_details.id,
		user_details.username,
		user_details.first_name,
		user_details.last_name,
		user_details.status,
		departments.dpt_name AS department
	FROM user_details
	LEFT JOIN departments ON user_details.department_id = departments.id`

	columnLabels := map[string]string{
		"id":         "ID",
		"username":   "Nome do Usuário",
		"first_name": "Nome",
		"last_name":  "Sobrenome",
		"status":     "Status",
		"department": "Departamento",
	}

	hiddenColumns := []string{"id"}

	handler.SetTableName("user_details")
	handler.SetQuery(query)
	handler.SetColumnLabels(columnLabels)
	handler.SetHiddenColumns(hiddenColumns)
	// handler.SetLinesPerPage(20)  // default is 15
	return nil
}
