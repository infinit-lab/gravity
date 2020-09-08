package database

import (
	"database/sql"
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"reflect"
	"strings"
	"sync"
)

type mysql struct {
	db        *sql.DB
	mutex     sync.Mutex
	isRunning bool
}

func (m *mysql) Begin() (*sql.Tx, error) {
	m.mutex.Lock()
	db := m.db
	m.mutex.Unlock()
	if db == nil {
		return nil, errors.New("The mysql is nil. ")
	}
	return db.Begin()
}

func (m *mysql) Prepare(query string) (stmt *sql.Stmt, err error) {
	m.mutex.Lock()
	db := m.db
	m.mutex.Unlock()
	if db == nil {
		return nil, errors.New("The mysql is nil. ")
	}
	stmt, err = db.Prepare(query)
	if err != nil {
		printer.Error(err)
		printer.Error(query)
	}
	return stmt, err
}

func (m *mysql) Exec(query string, args ...interface{}) (sql.Result, error) {
	m.mutex.Lock()
	db := m.db
	m.mutex.Unlock()
	if db == nil {
		return nil, errors.New("The mysql is nil. ")
	}
	result, err := db.Exec(query, args)
	if err != nil {
		printer.Error(err)
		printer.Error(query, args)
	}
	return result, err
}

func (m *mysql) Query(query string, args ...interface{}) (*sql.Rows, error) {
	m.mutex.Lock()
	db := m.db
	m.mutex.Unlock()
	if db == nil {
		return nil, errors.New("The mysql is nil. ")
	}
	row, err := db.Query(query, args)
	if err != nil {
		printer.Error(err)
		printer.Error(query, args)
	}
	return row, err
}

func (m *mysql) Close() {
	m.isRunning = false
}

func (m *mysql) NewTable(content interface{}, tableName string) (Table, error) {
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

	table := new(table)
	table.db = m
	table.tableName = tableName
	table.template = content
	return table, nil
}
