package database

import (
	"database/sql"
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"reflect"
	"strings"
	"sync"
)

type sqlite struct {
	db    *sql.DB
	mutex sync.Mutex
}

type sqliteTx struct {
	tx *sql.Tx
	db *sqlite
}

func (tx *sqliteTx) Commit() error {
	defer tx.db.mutex.Unlock()
	return tx.tx.Commit()
}

func (tx *sqliteTx) Rollback() error {
	defer tx.db.mutex.Unlock()
	return tx.tx.Rollback()
}

func (tx *sqliteTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := tx.tx.Exec(query, args...)
	if err != nil {
		printer.Error(err)
		printer.Error(query, args)
	}
	return result, err
}

func (d *sqlite) Begin() (Tx, error) {
	d.mutex.Lock()
	if d.db == nil {
		return nil, errors.New("The sqlite is nil. ")
	}
	tx := new(sqliteTx)
	var err error
	tx.tx, err = d.db.Begin()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	tx.db = d
	return tx, nil
}

func (d *sqlite) Prepare(query string) (stmt *sql.Stmt, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.db == nil {
		return nil, errors.New("The sqlite is nil. ")
	}
	stmt, err = d.db.Prepare(query)
	if err != nil {
		printer.Error(err)
		printer.Error(query)
	}
	return
}

func (d *sqlite) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.db == nil {
		return nil, errors.New("The sqlite is nil. ")
	}
	result, err = d.db.Exec(query, args...)
	if err != nil {
		printer.Error(err)
		printer.Error(query, args)
	}
	return
}
func (d *sqlite) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.db == nil {
		return nil, errors.New("The sqlite is nil. ")
	}
	rows, err = d.db.Query(query, args...)
	if err != nil {
		printer.Error(err)
		printer.Error(query, args)
	}
	return
}

func (d *sqlite) Close() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	_ = d.db.Close()
	d.db = nil
}

type column struct {
	cid     int64
	name    string
	t       string
	notnull int64
	dflt    sql.NullString
	pk      int64
}

type field struct {
	name    string
	t       string
	index   string
	dflt    string
	omit    string
	value   interface{}
	pointer interface{}
}

func parseStruct(value reflect.Value) (fields []*field) {
	if value.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < value.NumField(); i++ {
		switch value.Field(i).Kind() {
		case reflect.Struct:
			if value.Type().Field(i).Tag.Get("db_skip") == "true" {
				continue
			}
			fields = append(fields, parseStruct(value.Field(i))...)
		case reflect.Ptr, reflect.Array, reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer, reflect.Slice, reflect.Invalid:
			continue
		default:
			f := new(field)
			f.name = value.Type().Field(i).Tag.Get("db")
			f.t = value.Type().Field(i).Tag.Get("db_type")
			f.index = value.Type().Field(i).Tag.Get("db_index")
			f.dflt = value.Type().Field(i).Tag.Get("db_default")
			f.omit = value.Type().Field(i).Tag.Get("db_omit")
			if value.Field(i).CanInterface() {
				f.value = value.Field(i).Interface()
			}
			if value.Field(i).CanAddr() {
				f.pointer = value.Field(i).Addr().Interface()
			}
			if len(f.name) == 0 {
				continue
			}
			if len(f.t) == 0 {
				switch value.Type().Field(i).Type.Kind() {
				case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint8:
					f.t = "INTEGER"
				case reflect.Bool:
					f.t = "TINYINT"
				case reflect.String:
					f.t = "VARCHAR(256)"
				case reflect.Int64, reflect.Uint64:
					f.t = "BIGINT"
				default:
					continue
				}
			}
			fields = append(fields, f)
		}
	}
	return
}

func createColumnSql(field *field) string {
	query := "`" + field.name + "` " + field.t + " "
	if field.index == "primary" {
		query += "PRIMARY KEY AUTOINCREMENT"
	} else {
		query += "NOT NULL "
		if len(field.dflt) != 0 {
			query += "DEFAULT " + field.dflt
		}
	}
	return query
}

func (d *sqlite) createIndex(tableName string, field *field) error {
	if field.index == "index" || field.index == "unique" {
		query := "CREATE "
		if field.index == "unique" {
			query += "UNIQUE "
		}
		query += "INDEX " + field.name + "_" + tableName + "_index ON " + tableName + "(" + field.name + ")"
		_, err := d.Exec(query)
		if err != nil {
			printer.Error(err)
			return err
		}
	}
	return nil
}

func (d *sqlite) NewTable(content interface{}, tableName string) (Table, error) {
	v := reflect.ValueOf(content)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("The kind of content is not struct. ")
	}
	if tableName == "" {
		names := strings.Split(v.Type().Name(), ".")
		tableName = "t_" + strings.ToLower(names[len(names)-1])
	}

	fields := parseStruct(v)

	// get table info
	rows, err := d.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var columns []column
	for rows.Next() {
		var c column
		err := rows.Scan(&c.cid, &c.name, &c.t, &c.notnull, &c.dflt, &c.pk)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		columns = append(columns, c)
	}

	if len(columns) == 0 {
		query := "CREATE TABLE IF NOT EXISTS " + tableName + "("
		isFirst := true
		for _, field := range fields {
			if !isFirst {
				query += ","
			}
			q := createColumnSql(field)
			if len(q) != 0 {
				isFirst = false
			}
			query += q
		}
		query += ")"
		_, err := d.Exec(query)
		if err != nil {
			printer.Error(err)
			return nil, err
		}

		for _, field := range fields {
			err := d.createIndex(tableName, field)
			if err != nil {
				printer.Error(err)
				return nil, err
			}
		}
	} else {
		//check column
		for _, f := range fields {
			isFind := false
			for _, c := range columns {
				if c.name == f.name {
					isFind = true
					break
				}
			}
			if isFind {
				continue
			}
			query := "ALTER TABLE " + tableName + " ADD COLUMN " + createColumnSql(f)
			_, err := d.Exec(query)
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			err = d.createIndex(tableName, f)
			if err != nil {
				printer.Error(err)
				return nil, err
			}
		}
	}
	table := new(table)
	table.db = d
	table.tableName = tableName
	table.template = content
	return table, nil
}
