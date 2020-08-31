package sqlite

import (
	"database/sql"
	"github.com/infinit-lab/gravity/printer"
	_ "github.com/mattn/go-sqlite3"
)

type PrimaryKey struct {
	Id int `json:"id" db:"id" db_index:"primary" db_omit:"create,update"`
}

type Resource struct {
	PrimaryKey
	Name       string `json:"name" db:"name" db_type:"VARCHAR(256)" db_default:"''"`
	Creator    string `json:"creator" db:"creator" db_omit:"update" db_type:"VARCHAR(256)" db_default:"''"`
	Updater    string `json:"updater" db:"updater" db_omit:"create" db_type:"VARCHAR(256)" db_default:"''"`
	CreateTime string `json:"createTime" db:"createTime" db_omit:"create,update" db_type:"DATETIME" db_default:"CURRENT_TIMESTAMP"`
	UpdateTime string `json:"updateTime" db:"updateTime" db_omit:"create" db_type:"DATETIME" db_default:"CURRENT_TIMESTAMP"`
}

type Database interface {
	Prepare(query string) (stmt *sql.Stmt, err error)
	Exec(query string, args ...interface{}) (result sql.Result, err error)
	Query(query string, args ...interface{}) (rows *sql.Rows, err error)
	Close()

	NewTable(content interface{}) (table Table, err error)
}

type Table interface {
	Database() Database
	TableName() string

	GetList(whereSql string, args ...interface{}) (values []interface{}, err error)
	Get(whereSql string, args ...interface{}) (values interface{}, err error)
	Create(data interface{}) (sql.Result, error)
	Update(data interface{}, whereSql string, args ...interface{}) (sql.Result, error)
	Delete(whereSql string, args ...interface{}) (sql.Result, error)
}

func NewDatabase(path string) (Database, error) {
	sqlite, err := sql.Open("sqlite3", path)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	db := new(database)
	db.db = sqlite
	return db, nil
}
