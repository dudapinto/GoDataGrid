package setup

import (
	"datagrid/crud"
)

func UserDetailsSetup(handler *crud.CRUDHandler) error {

	query := "SELECT id, username, first_name, last_name, status FROM user_details"
	columnLabels := map[string]string{
		"username":   "Nome do Usuário",
		"id":         "ID",
		"first_name": "Nome",
		"last_name":  "Sobrenome",
		"status":     "Status",
	}
	handler.SetTableName("user_details")
	handler.SetQuery(query)
	handler.SetColumnLabels(columnLabels)
	handler.SetLinesPerPage(20)
	return nil
}
