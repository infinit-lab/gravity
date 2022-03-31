package model2

import (
	"errors"
	"fmt"
	"github.com/infinit-lab/gravity/event"
	mdl "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"sync"
)

type model struct {
	db         *gorm.DB
	data   IData
	tx         *gorm.DB
	mutex      sync.Mutex
	cache      map[string]IData
	cacheMutex sync.RWMutex
	eventList  []*event.Event
}

func (m *model) init(db *gorm.DB, data IData) error {
	m.db = db
	m.data = data
	m.cache = make(map[string]IData)
	err := m.db.AutoMigrate(data)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (m *model) Begin() {
	m.mutex.Lock()
	if m.tx != nil {
		m.tx.Rollback()
		m.tx = nil
	}
	m.tx = m.db.Begin()
	m.eventList = make([]*event.Event, 0)
}

func (m *model) Commit() {
	if m.tx != nil {
		m.tx.Commit()
		m.tx = nil
	}
	_ = event.PublishList(m.eventList)
	m.mutex.Unlock()
}

func (m *model) Rollback() {
	if m.tx != nil {
		m.tx.Rollback()
		m.tx = nil
	}
	m.eventList = nil
	m.mutex.Unlock()
}

func (m *model) getDB() *gorm.DB {
	if m.tx != nil {
		return m.tx
	}
	return m.db
}

func (m *model) deleteResource(code string) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	delete(m.cache, code)
}

func (m *model) getResource(code string) (IData, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	r, ok := m.cache[code]
	if !ok {
		return nil, false
	}
	return r, true
}

func (m *model) insertResource(r IData) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.cache[r.GetCode()] = r
}

func (m *model) notify(status string, r IData, context interface{}) {
	topic := m.data.Topic()
	if len(topic) == 0 {
		return
	}
	e := new(event.Event)
	e.Topic = topic
	e.Status = status
	e.Data = r
	e.Context = context
	if m.eventList != nil {
		m.eventList = append(m.eventList, e)
	} else {
		_ = event.Publish(e)
	}
}

func (m *model) Create(data IData, context interface{}) (string, error) {
	if m.tx == nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	code := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	data.SetCode(code)
	result := m.getDB().Create(data)
	if result.Error != nil {
		printer.Error(result.Error)
		return "", result.Error
	}

	temp, err := m.Get(code)
	if err != nil {
		printer.Error(err)
		return "", err
	}

	m.notify(mdl.StatusCreated, temp, context)

	return code, nil
}

func (m *model) Update(data IData, context interface{}) error {
	if m.tx == nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	temp, err := m.Get(data.GetCode())
	if err != nil {
		printer.Error(err)
		return err
	}
	data.SetID(temp.GetID())
	result := m.getDB().Updates(data)
	if result.Error != nil {
		printer.Error(err)
		return err
	}
	m.deleteResource(data.GetCode())

	temp, err = m.Get(data.GetCode())
	if err != nil {
		printer.Error(err)
		return err
	}
	m.notify(mdl.StatusUpdated, temp, context)
	return nil
}

func (m *model) Delete(code string, context interface{}) error {
	if m.tx == nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	r, err := m.Get(code)
	if err != nil {
		printer.Error(err)
		return err
	}
	result := m.getDB().Delete(r)
	if result.Error != nil {
		printer.Error(err)
		return err
	}
	m.deleteResource(code)
	m.notify(mdl.StatusDeleted, r, context)
	return nil
}

func (m *model) newResource() IData {
	t := reflect.TypeOf(m.data)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface().(IData)
	}
	return reflect.New(t).Interface().(IData)
}

func (m *model) Get(code string) (IData, error) {
	r, ok := m.getResource(code)
	if ok {
		return r, nil
	}
	r = m.newResource()
	db := m.getDB()
	result := db.Where("`code` = ?", code).Find(r)
	if result.Error != nil {
		printer.Error(result.Error)
		return nil, result.Error
	}
	if r.GetID() == 0 {
		return nil, errors.New(fmt.Sprintf("Not Found %s", code))
	}
	m.insertResource(r)
	return r, nil
}

func (m *model) GetList(query string, conditions ...interface{}) (interface{}, error) {
	slice := reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(m.data)), 0, 0).Type()).Elem().Interface()
	var result *gorm.DB
	if len(query) == 0 {
		result = m.getDB().Find(&slice)
	} else {
		result = m.getDB().Where(query, conditions...).Find(&slice)
	}
	if result.Error != nil {
		printer.Error(result.Error)
		return nil, result.Error
	}
	return slice, nil
}

