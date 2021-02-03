package model

import (
	"database/sql"
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/snow_flack"
	"strconv"
	"sync"
)

type model struct {
	table database.Table
	topic string
	cacheDatabase database.Database
	cacheTable database.Table
	mutex sync.RWMutex
	getLayer func(resource interface{})
	notifyLayer func(e *event.Event)
}

func (m *model)init(db database.Database, resource interface{}, topic string, isCache bool, tableName string) error {
	var err error
	m.table, err = db.NewTable(resource, tableName)
	if err != nil {
		printer.Error(err)
		return err
	}
	m.topic = topic
	if isCache {
		m.cacheDatabase, err = database.NewDatabase("sqlite3", ":memory:?shared=false")
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
	}
	return nil
}

func (m *model)Table() database.Table {
	return m.table
}

func (m *model)GetList(whereSql string, args ...interface{}) ([]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	var values []interface{}
	var err error
	if m.cacheTable != nil {
		values, err = m.cacheTable.GetList(whereSql, args...)
	} else {
		values, err = m.table.GetList(whereSql, args...)
	}
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	if m.getLayer != nil {
		for _, value := range values {
			m.getLayer(value)
		}
	}
	return values, nil
}

func (m *model)Get(whereSql string, args ...interface{}) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	var value interface{}
	var err error
	if m.cacheTable != nil {
		value, err = m.cacheTable.Get(whereSql, args...)
	} else {
		value, err = m.table.Get(whereSql, args...)
	}
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	if m.getLayer != nil {
		m.getLayer(value)
	}
	return value, nil
}

func (m *model)GetByCode(code string) (interface{}, error) {
	return m.Get("WHERE `code` = ?", code)
}

func (m *model)Create(resource Code, context interface{}) (string, error) {
	if resource.GetCode() == "" {
		code, err := snow_flack.NextId()
		if err != nil {
			printer.Error(err)
			return "", err
		}
		resource.SetCode(strconv.FormatInt(code, 16))
	}
	err := func() error {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		_, err := m.table.Create(resource)
		if err != nil {
			printer.Error(err)
			return err
		}
		if m.cacheTable != nil {
			_, err = m.cacheTable.Create(resource)
			if err != nil {
				printer.Error(err)
				return err
			}
		}
		return nil
	}()
	if err != nil {
		printer.Error(err)
		return "", err
	}
	value, err := m.GetByCode(resource.GetCode())
	if err != nil {
		printer.Error(err)
		return "", err
	}
	if len(m.topic) != 0 {
		e := new(event.Event)
		e.Topic = m.topic
		e.Status = StatusCreated
		e.Data = value
		e.Context = context
		if m.notifyLayer != nil {
			m.notifyLayer(e)
		}
		_ = event.Publish(e)
	}
	return resource.GetCode(), nil
}

func (m *model)Update(resource interface{}, context interface{}, whereSql string, args ...interface{}) error {
	ret, err := func() (sql.Result, error) {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		ret, err := m.table.Update(resource, whereSql, args...)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		if m.cacheTable != nil {
			_, err = m.cacheTable.Update(resource, whereSql, args...)
			if err != nil {
				printer.Error(err)
				return nil, err
			}
		}
		return ret, nil
	}()
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
	values, err := m.GetList(whereSql, args...)
	for _, value := range values {
		if len(m.topic) != 0 {
			e := new(event.Event)
			e.Topic = m.topic
			e.Status = StatusUpdated
			e.Data = value
			e.Context = context
			if m.notifyLayer != nil {
				m.notifyLayer(e)
			}
			_ = event.Publish(e)
		}
	}
	return nil
}

func (m *model)UpdateByCode(resource Code, context interface{}) error {
	return m.Update(resource, context, "WHERE `code` = ?", resource.GetCode())
}

func (m *model)Delete(context interface{}, whereSql string, args ...interface{}) error {
	values, err := m.GetList(whereSql, args...)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = func() error {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		_, err = m.table.Delete(whereSql, args...)
		if err != nil {
			printer.Error(err)
			return err
		}
		if m.cacheTable != nil {
			_, err = m.cacheTable.Delete(whereSql, args...)
			if err != nil {
				printer.Error(err)
				return err
			}
		}
		return nil
	}()
	if err != nil {
		printer.Error(err)
		return err
	}
	for _, value := range values {
		if len(m.topic) != 0 {
			e := new(event.Event)
			e.Topic = m.topic
			e.Status = StatusDeleted
			e.Data = value
			e.Context = context
			if m.notifyLayer != nil {
				m.notifyLayer(e)
			}
			_ = event.Publish(e)
		}
	}
	return nil
}

func (m *model)DeleteByCode(code string, context interface{}) error {
	return m.Delete(context, "WHERE `code` = ?", code)
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
	values, err := m.table.GetList("")
	if err != nil {
		printer.Error(err)
		return err
	}
	tx, err := m.cacheDatabase.Begin()
	if err != nil {
		printer.Error(err)
		return err
	}
	var query string
	var args []interface{}
	for _, value := range values {
		query, args, err = m.cacheTable.CreateSql(value)
		if err != nil {
			printer.Error(err)
			break
		}
		_, err = tx.Exec(query, args...)
		if err != nil {
			printer.Error(err)
			break
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	} else {
		_ = tx.Commit()
	}
	return nil
}

func (m *model)SetBeforeGetLayer(layer func(resource interface{})) {
	m.getLayer = layer
}

func (m *model)SetBeforeNotifyLayer(layer func(e *event.Event)) {
	m.notifyLayer = layer
}

