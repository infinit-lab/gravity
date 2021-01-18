package model2

import (
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/printer"
)

type Model interface {
	Table() database.Table

	GetList(whereSql string, args ...interface{}) ([]interface{}, error)
	Get(whereSql string, args ...interface{}) (interface{}, error)
	Create(resource interface{}, context interface{}) error
	Update(resource interface{}, context interface{}, whereSql string, args ...interface{}) error
	Delete(context interface{}, whereSql string, args ...interface{}) error
	Sync() error
	SyncSingle(whereSql string, args ...interface{}) (interface{}, error)
}

func New(db database.Database, resource interface{}, topic string, isCache bool, tableName string) (Model, error) {
	m := new(model)
	err := m.init(db, resource, topic, isCache, tableName)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return m, nil
}
