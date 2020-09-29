package database

import (
	"database/sql"
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"reflect"
	"strings"
	"time"
)

type table struct {
	db        Database
	tableName string
	template  interface{}
}

func (t *table) Database() Database {
	return t.db
}

func (t *table) TableName() string {
	return t.tableName
}

func (t *table) GetList(whereSql string, args ...interface{}) (values []interface{}, err error) {
	fields, err := parseStructWithOmit(t.template, "get")
	query := "SELECT "
	for i, f := range fields {
		query += "`" + f.name + "`"
		if i != len(fields)-1 {
			query += ", "
		}
	}
	query += " FROM " + t.tableName + " "
	query += whereSql
	rows, err := t.db.Query(query, args...)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		typ := reflect.TypeOf(t.template)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		value := reflect.New(typ).Interface()
		fields, err := parseStructWithOmit(value, "get")
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		var list []interface{}
		for _, f := range fields {
			list = append(list, f.pointer)
		}
		err = rows.Scan(list...)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		for _, f := range fields {
			if f.t == "DATETIME" {
				str := *(f.pointer.(*string))
				str = strings.ReplaceAll(str, "T", " ")
				str = strings.ReplaceAll(str, "Z", "")
				t, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.UTC)
				if err != nil {
					printer.Error(err)
					continue
				}
				*(f.pointer.(*string)) = t.Local().Format("2006-01-02 15:04:05")
			}
		}
		values = append(values, value)
	}
	return
}

func (t *table) Get(whereSql string, args ...interface{}) (value interface{}, err error) {
	values, err := t.GetList(whereSql, args...)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, errors.New("Not found. ")
	}
	return values[0], nil
}

func (t *table) Create(data interface{}) (sql.Result, error) {
	fields, err := parseStructWithOmit(data, "create")
	if err != nil {
		printer.Trace(err)
		return nil, err
	}
	query := "INSERT INTO " + t.tableName + " ("
	for i, f := range fields {
		query += "`" + f.name + "`"
		if i != len(fields)-1 {
			query += ", "
		}
	}
	query += ") VALUES ("
	var values []interface{}
	for i, f := range fields {
		query += "?"
		if i != len(fields)-1 {
			query += ", "
		}
		values = append(values, f.value)
	}
	query += ")"
	ret, err := t.db.Exec(query, values...)
	return ret, err
}

func (t *table) Update(data interface{}, whereSql string, args ...interface{}) (sql.Result, error) {
	fields, err := parseStructWithOmit(data, "update")
	if err != nil {
		printer.Trace(err)
		return nil, err
	}
	var values []interface{}
	query := "UPDATE " + t.tableName + " SET "
	for i, f := range fields {
		query += "`" + f.name + "` = ?"
		if i != len(fields)-1 {
			query += ", "
		}
		values = append(values, f.value)
	}
	query += whereSql
	values = append(values, args...)
	ret, err := t.db.Exec(query, values...)
	return ret, err
}

func (t *table) Delete(whereSql string, args ...interface{}) (sql.Result, error) {
	query := "DELETE FROM " + t.tableName + " "
	query += whereSql
	ret, err := t.db.Exec(query, args...)
	return ret, err
}

func parseStructWithOmit(data interface{}, omit string) ([]*field, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("The kind of data is not struct. ")
	}
	fields := parseStruct(v)
	var tempFields []*field
	for _, f := range fields {
		if strings.Contains(f.omit, omit) {
			continue
		}
		tempFields = append(tempFields, f)
	}
	return tempFields, nil
}

func generateWhereSql(keys map[string]string) (string, []interface{}) {
	if len(keys) == 0 {
		return "", nil
	}
	var values []interface{}
	query := " WHERE "
	isFirst := true
	for key, value := range keys {
		if !isFirst {
			query += " AND "
		}
		isFirst = false
		query += "`" + key + "` = ?"
		values = append(values, value)
	}
	return query, values
}
