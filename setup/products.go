package setup

import "datagrid/crud"

func ProductsSetup(handler *crud.CRUDHandler) error {
	handler.SetTableName("products")
	query := "SELECT id, name, description, price FROM products"
	columnLabels := map[string]string{
		"id":          "Product ID",
		"name":        "Name",
		"description": "Description",
		"price":       "Price",
	}
	handler.SetQuery(query)
	handler.SetColumnLabels(columnLabels)
	handler.SetLinesPerPage(20)
	return nil
}
