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

func (e inExpr) ToSql() (sql string, args []any, err error) {
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
				sql = fmt.Sprintf("%s=?", e.column)
			} else {
				args = []any{v}
				sql = fmt.Sprintf("%s=ANY(?)", e.column)
			}
		} else {
			args = []any{v}
			sql = fmt.Sprintf("%s=?", e.column)
		}
	}

	return sql, args, err
}

// notInExpr helps to use NOT IN in SQL query
type notInExpr inExpr

// NotIn allows to use NOT IN in SQL query
// Ex: SelectBuilder.Where(NotIn("id", 1, 2, 3))
func NotIn(column string, e any) notInExpr {
	return notInExpr{column, e}
}

func (e notInExpr) ToSql() (sql string, args []any, err error) {
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
				sql = fmt.Sprintf("%s<>?", e.column)
			} else {
				args = []any{v}
				sql = fmt.Sprintf("%s<>ALL(?)", e.column)
			}
		} else {
			args = []any{v}
			sql = fmt.Sprintf("%s<>?", e.column)
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
