package setup

import "datagrid/crud"

func ProductsSetup(handler *crud.CRUDHandler) error {
	handler.SetTableName("products")
	query := "SELECT product_id, name, description, price FROM products"
	columnLabels := map[string]string{
		"product_id":  "Product ID",
		"name":        "Name",
		"description": "Description",
		"price":       "Price",
	}
	handler.SetQuery(query)
	handler.SetColumnLabels(columnLabels)
	handler.SetLinesPerPage(20)
	return nil
}
