package model

import (
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"sync"
)

type Id interface {
	GetId() int
	SetId(id int)
}

type PrimaryKey struct {
	database.PrimaryKey
}

type Resource struct {
	database.Resource
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
	TopicSync     string = "sync"
	StatusCreated string = "created"
	StatusUpdated string = "updated"
	StatusDeleted string = "deleted"
)

type Model interface {
	Table() database.Table

	GetList() ([]interface{}, error)
	GetListWithWhereSql(whereSql string, args ...interface{}) ([]interface{}, error)
	Get(id int) (interface{}, error)
	Create(resource Id, context interface{}) (int, error)
	Update(resource Id, context interface{}) error
	Delete(id int, context interface{}) error
	Sync() error
	SyncSingle(id int) (interface{}, error)
	SetBeforeInsertLayer(layer func(id int, resource interface{}))
	SetBeforeEraseLayer(layer func(id int, resource interface{}))
	SetBeforeNotifyLayer(layer func(id int, e *event.Event))
}

func New(db database.Database, resource Id, topic string, isCache bool, tableName string) (Model, error) {
	tb, err := db.NewTable(resource, tableName)
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
	model.subscriber, _ = event.Subscribe(TopicSync, func(e *event.Event) {
		_ = model.Sync()
	})

	return model, nil
}

type model struct {
	table       database.Table
	topic       string
	cache       Cache
	mutex       sync.RWMutex
	subscriber  event.Subscriber
	insertLayer func(id int, resource interface{})
	eraseLayer  func(id int, resource interface{})
	notifyLayer func(id int, e *event.Event)
}

func (m *model) Table() database.Table {
	return m.table
}

func (m *model) getList() ([]interface{}, error) {
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
			m.cache.Insert(v.GetId(), v, m.insertLayer)
		}
	}
	return values, err
}

func (m *model) GetList() ([]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.getList()
}

func (m *model) getIdList(whereSql string, args ...interface{}) ([]int, error) {
	rows, err := m.Table().Database().Query("SELECT `id` FROM " + m.Table().TableName() + " " + whereSql, args...)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var idList []int
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		idList = append(idList, id)
	}
	return idList, nil
}

func (m *model) GetListWithWhereSql(whereSql string, args ...interface{}) ([]interface{}, error) {
	if m.cache != nil {
		idList, err := m.getIdList(whereSql, args...)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		m.mutex.RLock()
		defer m.mutex.RUnlock()
		var list []interface{}
		for _, id := range idList {
			i, err := m.get(id)
			if err != nil {
				printer.Error(err)
			}
			list = append(list, i)
		}
		return list, nil
	} else {
		return m.Table().GetList(whereSql, args...)
	}
}

func (m *model) get(id int) (interface{}, error) {
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
		m.cache.Insert(id, value, m.insertLayer)
	}
	return value, nil
}

func (m *model) Get(id int) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.get(id)
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
	if len(m.topic) != 0 {
		e := new(event.Event)
		e.Topic = m.topic
		e.Status = StatusCreated
		e.Data = value
		e.Context = context
		if m.notifyLayer != nil {
			m.notifyLayer(int(id), e)
		}
		_ = event.Publish(e)
	}
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
	value, err := m.SyncSingle(resource.GetId())
	if len(m.topic) != 0 {
		e := new(event.Event)
		e.Topic = m.topic
		e.Status = StatusUpdated
		e.Data = value
		e.Context = context
		if m.notifyLayer != nil {
			m.notifyLayer(resource.GetId(), e)
		}
		_ = event.Publish(e)
	}
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
		m.cache.Erase(id, m.eraseLayer)
	}
	if len(m.topic) != 0 {
		e := new(event.Event)
		e.Topic = m.topic
		e.Status = StatusDeleted
		e.Data = value
		e.Context = context
		if m.notifyLayer != nil {
			m.notifyLayer(id, e)
		}
		_ = event.Publish(e)
	}
	return nil
}

func (m *model) Sync() error {
	if m.cache == nil {
		return nil
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cache.Clear()
	_, err := m.getList()
	return err
}

func (m *model) SyncSingle(id int) (interface{}, error) {
	if m.cache == nil {
		return m.Get(id)
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cache.Erase(id, m.eraseLayer)
	return m.get(id)
}

func (m *model) SetBeforeInsertLayer(layer func(id int, resource interface{})) {
	m.insertLayer = layer
	_ = m.Sync()
}

func (m *model) SetBeforeEraseLayer(layer func(id int, resource interface{})) {
	m.eraseLayer = layer
	_ = m.Sync()
}

func (m *model) SetBeforeNotifyLayer(layer func(id int, e *event.Event)) {
	m.notifyLayer = layer
}
