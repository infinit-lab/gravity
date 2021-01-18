package model2

import (
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"sync"
)

type model struct {
	table database.Table
	topic string
	cacheDatabase database.Database
	cacheTable database.Table
	mutex sync.RWMutex
}

const (
	StatusCreated string = "created"
	StatusUpdated string = "updated"
	StatusDeleted string = "deleted"
)

func (m *model)init(db database.Database, resource interface{}, topic string, isCache bool, tableName string) error {
	var err error
	m.table, err = db.NewTable(resource, tableName)
	if err != nil {
		printer.Error(err)
		return err
	}
	m.topic = topic
	m.cacheDatabase, err = database.NewDatabase("sqlite3", ":memory:")
	if err != nil {
		printer.Error(err)
		return err
	}
	m.cacheTable, err = m.cacheDatabase.NewTable(resource, tableName)
	if err != nil {
		m.cacheDatabase.Close()
		m.cacheDatabase = nil
		printer.Error(err)
		return err
	}
	err = m.Sync()
	if err != nil {
		m.cacheDatabase.Close()
		m.cacheDatabase = nil
		printer.Error(err)
		return err
	}
	return nil
}

func (m *model)Table() database.Table {
	return m.table
}

func (m *model)getList(whereSql string, args ...interface{}) ([]interface{}, error) {
	var values []interface{}
	var err error
	if m.cacheTable != nil {
		values, err = m.cacheTable.GetList(whereSql, args...)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		if len(values) != 0 {
			printer.Trace("Get from cache")
		}
	}
	if len(values) == 0 {
		values, err = m.table.GetList(whereSql, args...)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		for _, value := range values {
			_, err := m.cacheTable.Create(value)
			if err != nil {
				printer.Error(err)
				_, _ = m.cacheTable.Delete("")
				return nil, err
			}
		}
	}
	return values, nil
}

func (m *model)get(whereSql string, args ...interface{}) (interface{}, error) {
	var value interface{}
	var err error
	if m.cacheTable != nil {
		value, _ = m.cacheTable.Get(whereSql, args...)
		if value != nil {
			printer.Trace("Get from cache.")
			return value, nil
		}
	}
	value, err = m.table.Get(whereSql, args...)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	_, _ = m.cacheTable.Create(value)
	return value, nil
}

func (m *model)GetList(whereSql string, args ...interface{}) ([]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.getList(whereSql, args...)
}

func (m *model)Get(whereSql string, args ...interface{}) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.get(whereSql, args...)
}

func (m *model)Create(resource interface{}, context interface{}) error {
	ret, err := m.table.Create(resource)
	if err != nil {
		printer.Error(err)
		return err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		printer.Error(err)
		return err
	}
	value, err := m.Get("WHERE `id` = ?", id)
	if err != nil {
		printer.Error(err)
		return err
	}
	if len(m.topic) != 0 {
		e := new(event.Event)
		e.Topic = m.topic
		e.Status = StatusCreated
		e.Data = value
		e.Context = context
		/*
		if m.notifyLayer != nil {
			m.notifyLayer(int(id), e)
		}
		*/
		_ = event.Publish(e)
	}
	return nil
}

func (m *model)Update(resource interface{}, context interface{}, whereSql string, args ...interface{}) error {
	ret, err := m.table.Update(resource, whereSql, args...)
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
	value, err := m.SyncSingle(whereSql, args...)
	if len(m.topic) != 0 {
		e := new(event.Event)
		e.Topic = m.topic
		e.Status = StatusUpdated
		e.Data = value
		e.Context = context
		/*
		if m.notifyLayer != nil {
			m.notifyLayer(resource.GetId(), e)
		}
		*/
		_ = event.Publish(e)
	}
	return nil
}

func (m *model)Delete(context interface{}, whereSql string, args ...interface{}) error {
	value, err := m.Get(whereSql, args...)
	if err != nil {
		printer.Error(err)
		return err
	}
	_, err = m.table.Delete(whereSql, args...)
	if err != nil {
		printer.Error(err)
		return err
	}
	if m.cacheTable != nil {
		_, _ = m.table.Delete(whereSql, args...)
	}
	if len(m.topic) != 0 {
		e := new(event.Event)
		e.Topic = m.topic
		e.Status = StatusDeleted
		e.Data = value
		e.Context = context
		/*
		if m.notifyLayer != nil {
			m.notifyLayer(id, e)
		}
		*/
		_ = event.Publish(e)
	}
	return nil
}

func (m *model)Sync() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.cacheTable == nil {
		return nil
	}
	_, err := m.cacheTable.Delete("")
	if err != nil {
		printer.Error(err)
		return err
	}
	_, err = m.getList("")
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (m *model)SyncSingle(whereSql string, args ...interface{}) (interface{}, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.cacheTable != nil {
		_, err := m.cacheTable.Delete(whereSql, args...)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
	}
	return m.get(whereSql, args...)
}

