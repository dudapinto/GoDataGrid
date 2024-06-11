package setup

import "datagrid/crud"

func OrdersSetup(handler *crud.CRUDHandler) error {
	handler.SetTableName("orders")
	query := "SELECT order_id, user_id, product_id, quantity, order_date FROM orders"
	columnLabels := map[string]string{
		"order_id":   "Order ID",
		"user_id":    "User ID",
		"product_id": "Product ID",
		"quantity":   "Quantity",
		"order_date": "Order Date",
	}
	handler.SetQuery(query)
	handler.SetColumnLabels(columnLabels)
	handler.SetLinesPerPage(20)
	return nil
}
