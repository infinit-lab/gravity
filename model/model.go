package model

import (
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
)

type Code interface {
	GetCode() string
	SetCode(code string)
}

type PrimaryKey struct {
	database.PrimaryKey
	Code string `json:"code" db:"code" db_type:"VARCHAR(64)" db_index:"unique" db_omit:"update" db_default:"''"`
}

func (p *PrimaryKey) GetCode() string {
	return p.Code
}

func (p *PrimaryKey) SetCode(code string) {
	p.Code = code
}

type Resource struct {
	database.Resource
	Code string `json:"code" db:"code" db_type:"VARCHAR(64)" db_index:"unique" db_omit:"update" db_default:"''"`
}

func (r *Resource) GetCode() string {
	return r.Code
}

func (r *Resource) SetCode(code string) {
	r.Code = code
}

const (
	TopicSync     string = "sync"
	StatusCreated string = "created"
	StatusUpdated string = "updated"
	StatusDeleted string = "deleted"
)

type Model interface {
	Table() database.Table

	GetList(whereSql string, args ...interface{}) ([]interface{}, error)
	Get(whereSql string, args ...interface{}) (interface{}, error)
	GetByCode(code string) (interface{}, error)
	Create(resource Code, context interface{}) (string, error)
	Update(resource interface{}, context interface{}, whereSql string, args ...interface{}) error
	UpdateByCode(resource Code, context interface{}) error
	Delete(context interface{}, whereSql string, args ...interface{}) error
	DeleteByCode(code string, context interface{}) error
	Sync() error
	SetBeforeGetLayer(layer func(resource interface{}))
	SetBeforeNotifyLayer(layer func(resource interface{}))
}

func New(db database.Database, resource Code, topic string, isCache bool, tableName string) (Model, error) {
	m := new(model)
	err := m.init(db, resource, topic, isCache, tableName)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	subscriber, _ := event.Subscribe(TopicSync)
	go func() {
		for {
			_ = <-subscriber.Event()
			err := m.Sync()
			if err != nil {
				printer.Error(err)
				continue
			}
		}
	}()
	return m, nil
}
