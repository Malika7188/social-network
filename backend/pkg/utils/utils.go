package utils

import (
	"database/sql"
	"fmt"
)

// checks nullable stirngs
func NullableString(ns sql.NullString) interface{} {
	fmt.Println("=============NullableString:===================", ns)
	if ns.Valid {
		fmt.Println("NullableString:", ns.String)
		return ns.String
	}
	return nil
}

func NullableInt64(ns sql.NullInt64) interface{} {
	if ns.Valid {
		return ns.Int64
	}
	return nil
}
