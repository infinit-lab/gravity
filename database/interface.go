package database

import (
	"database/sql"
	"errors"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type PrimaryKey struct {
	Id int `json:"id" db:"id" db_index:"primary" db_omit:"create,update"`
}

type Resource struct {
	PrimaryKey
	Name       string `json:"name" db:"name" db_type:"VARCHAR(256)" db_default:"''"`
	Creator    string `json:"creator" db:"creator" db_omit:"update" db_type:"VARCHAR(256)" db_default:"''"`
	Updater    string `json:"updater" db:"updater" db_type:"VARCHAR(256)" db_default:"''"`
	CreateTime string `json:"createTime" db:"createTime" db_omit:"create,update" db_type:"DATETIME" db_default:"CURRENT_TIMESTAMP"`
	UpdateTime string `json:"updateTime" db:"updateTime" db_omit:"create" db_type:"DATETIME" db_default:"CURRENT_TIMESTAMP"`
}

type Tx interface {
	Commit() error
	Rollback() error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Database interface {
	Begin() (Tx, error)
	Prepare(query string) (stmt *sql.Stmt, err error)
	Exec(query string, args ...interface{}) (result sql.Result, err error)
	Query(query string, args ...interface{}) (rows *sql.Rows, err error)
	Close()

	NewTable(content interface{}, tableName string) (table Table, err error)
}

type Table interface {
	Database() Database
	TableName() string

	GetList(whereSql string, args ...interface{}) (values []interface{}, err error)
	Get(whereSql string, args ...interface{}) (values interface{}, err error)
	Create(data interface{}) (sql.Result, error)
	CreateSql(data interface{}) (string, []interface{}, error)
	Update(data interface{}, whereSql string, args ...interface{}) (sql.Result, error)
	Delete(whereSql string, args ...interface{}) (sql.Result, error)
}

func NewDatabase(driverName string, dataSourceName string) (Database, error) {
	if driverName == "sqlite3" {
		s, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		db := new(sqlite)
		db.db = s
		return db, nil
	} else if driverName == "mysql" {
		m := new(mysql)
		m.isRunning = true
		go func() {
			for m.isRunning {
				var err error
				m.mutex.Lock()
				m.db, err = sql.Open(driverName, dataSourceName)
				m.mutex.Unlock()
				if err != nil {
					printer.Error(err)
				} else {
					maxOpenConns := config.GetInt("mysql.maxOpenConns")
					if maxOpenConns == 0 {
						maxOpenConns = 20
					}
					maxIdleConns := config.GetInt("mysql.maxIdleConns")
					if maxIdleConns == 0 {
						maxIdleConns = 5
					}
					maxLifetime := config.GetInt("mysql.maxLifetime")
					if maxLifetime == 0 {
						maxLifetime = 120
					}
					m.db.SetMaxOpenConns(maxOpenConns)
					m.db.SetMaxIdleConns(maxIdleConns)
					m.db.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
					for {
						err := m.db.Ping()
						if err != nil {
							printer.Error(err)
							m.mutex.Lock()
							_ = m.db.Close()
							m.db = nil
							m.mutex.Unlock()
							break
						}
					}
				}
				time.Sleep(5 * time.Second)
			}
			if m.db != nil {
				_ = m.db.Close()
			}
		}()
		return m, nil
	}
	return nil, errors.New("Not supported. ")
}
