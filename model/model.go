package model

import (
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/sqlite"
)

type Id interface {
	GetId() int
	SetId(id int)
}

type PrimaryKey struct {
	sqlite.PrimaryKey
}

type Resource struct {
	sqlite.Resource
}

func (p *PrimaryKey) GetId() int {
	return p.Id
}

func (p *PrimaryKey) SetId(id int) {
	p.Id = id
}

func (r *Resource) GetId() int {
	return r.Id
}

func (r *Resource) SetId(id int) {
	r.Id = id
}

const (
	StatusCreated string = "created"
	StatusUpdated string = "updated"
	StatusDeleted string = "deleted"
)

type Model interface {
	Table() sqlite.Table

	GetList() ([]interface{}, error)
	Get(id int) (interface{}, error)
	Create(resource Id, context interface{}) (int, error)
	Update(resource Id, context interface{}) error
	Delete(id int, context interface{}) error
	Sync() error
}

func New(db sqlite.Database, resource Id, topic string, isCache bool) (Model, error) {
	tb, err := db.NewTable(resource)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	model := new(model)
	model.table = tb
	model.topic = topic
	if isCache {
		model.cache = NewCache()
		_, _ = model.GetList()
	} else {
		model.cache = nil
	}
	return model, nil
}

type model struct {
	table sqlite.Table
	topic string

	cache Cache
}

func (m *model) Table() sqlite.Table {
	return m.table
}

func (m *model) GetList() ([]interface{}, error) {
	if m.cache != nil {
		values := m.cache.GetAll()
		if len(values) != 0 {
			return values, nil
		}
	}
	values, err := m.table.GetList("")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	if m.cache != nil {
		m.cache.Clear()
		for _, value := range values {
			v, ok := value.(Id)
			if !ok {
				continue
			}
			m.cache.Insert(v.GetId(), v)
		}
	}
	return values, err
}

func (m *model) Get(id int) (interface{}, error) {
	if m.cache != nil {
		value, ok := m.cache.Get(id)
		if ok {
			return value, nil
		}
	}
	value, err := m.table.Get("WHERE `id` = ? LIMIT 1", id)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	if m.cache != nil {
		m.cache.Insert(id, value)
	}
	return value, nil
}

func (m *model) Create(resource Id, context interface{}) (int, error) {
	ret, err := m.table.Create(resource)
	if err != nil {
		printer.Error(err)
		return 0, err
	}

	id, err := ret.LastInsertId()
	if err != nil {
		printer.Error(err)
		return 0, err
	}
	value, err := m.Get(int(id))
	if err != nil {
		printer.Error(err)
		return 0, err
	}
	e := new(event.Event)
	e.Topic = m.topic
	e.Status = StatusCreated
	e.Data = value
	e.Context = context
	_ = event.Publish(e)
	return int(id), nil
}

func (m *model) Update(resource Id, context interface{}) error {
	ret, err := m.table.Update(resource, "WHERE `id` = ?", resource.GetId())
	if err != nil {
		printer.Error(err)
		return err
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		printer.Error(err)
		return err
	}
	if rows == 0 {
		return nil
	}
	if m.cache != nil {
		m.cache.Erase(resource.GetId())
	}
	value, err := m.Get(resource.GetId())
	e := new(event.Event)
	e.Topic = m.topic
	e.Status = StatusUpdated
	e.Data = value
	e.Context = context
	_ = event.Publish(e)
	return nil
}

func (m *model) Delete(id int, context interface{}) error {
	value, err := m.Get(id)
	if err != nil {
		printer.Error(err)
		return err
	}
	_, err = m.table.Delete("WHERE `id` = ?", id)
	if err != nil {
		printer.Error(err)
		return err
	}
	if m.cache != nil {
		m.cache.Erase(id)
	}
	e := new(event.Event)
	e.Topic = m.topic
	e.Status = StatusDeleted
	e.Data = value
	e.Context = context
	_ = event.Publish(e)
	return nil
}

func (m *model) Sync() error {
	if m.cache == nil {
		return nil
	}
	m.cache.Clear()
	_, err := m.GetList()
	return err
}
