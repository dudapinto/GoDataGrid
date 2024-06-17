package setup

import "datagrid/crud"

func DepartmentsSetup(handler *crud.CRUDHandler) error {

	handler.SetTableName("departments")

	return nil

}
