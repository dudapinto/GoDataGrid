package loaders

import (
	"fmt"
	"log"
	"strings"

	"datagrid/crud"
	"datagrid/setup"
)

func LoadHandlerSetupFunc(tableName string) func(*crud.CRUDHandler) error {

	log.Println("(in LoadHandlerSetupFunc) tableName = ", tableName)
	fmt.Println("====================================================")
	switch strings.ToLower(tableName) {
	case "user_details":
		return setup.UserDetailsSetup
	case "orders":
		return setup.OrdersSetup
	case "products":
		return setup.ProductsSetup
	default:
		return nil
	}
}
