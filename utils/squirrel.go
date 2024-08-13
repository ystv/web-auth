package utils

import (
	"fmt"
	"reflect"

	"database/sql/driver"

	sq "github.com/Masterminds/squirrel"
)

func PSQL() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

// inExpr helps to use IN in SQL query
type inExpr struct {
	column string
	expr   any
}

//nolint:revive
func (e inExpr) ToSql() (sql string, args []interface{}, err error) {
	switch v := e.expr.(type) {
	case sq.Sqlizer:
		sql, args, err = v.ToSql()
		if err == nil && sql != "" {
			sql = fmt.Sprintf("%s IN (%s)", e.column, sql)
		}
	default:
		if isListType(v) {
			if reflect.ValueOf(v).Len() == 0 {
				return "", nil, nil
			}

			if reflect.ValueOf(v).Len() == 1 {
				args = []any{reflect.ValueOf(v).Index(0).Interface()}
				sql = e.column + "=?"
			} else {
				args = []any{v}
				sql = e.column + "=ANY(?)"
			}
		} else {
			args = []any{v}
			sql = e.column + "=?"
		}
	}

	return sql, args, err
}

// NotInExpr helps to use NOT IN in SQL query
type NotInExpr inExpr

// NotIn allows to use NOT IN in SQL query
// Ex: SelectBuilder.Where(NotIn("id", 1, 2, 3))
func NotIn(column string, e any) NotInExpr {
	return NotInExpr{column, e}
}

//nolint:revive
func (e NotInExpr) ToSql() (sql string, args []interface{}, err error) {
	switch v := e.expr.(type) {
	case sq.Sqlizer:
		sql, args, err = v.ToSql()
		if err == nil && sql != "" {
			sql = fmt.Sprintf("%s NOT IN (%s)", e.column, sql)
		}
	default:
		if isListType(v) {
			if reflect.ValueOf(v).Len() == 0 {
				return "", nil, nil
			}

			if reflect.ValueOf(v).Len() == 1 {
				args = []any{reflect.ValueOf(v).Index(0).Interface()}
				sql = e.column + "<>?"
			} else {
				args = []any{v}
				sql = e.column + "<>ALL(?)"
			}
		} else {
			args = []any{v}
			sql = e.column + "<>?"
		}
	}

	return sql, args, err
}

func isListType(val any) bool {
	if driver.IsValue(val) {
		return false
	}
	valVal := reflect.ValueOf(val)
	return valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice
}
