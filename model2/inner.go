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
	resource   IResource
	tx         *gorm.DB
	mutex      sync.Mutex
	cache      map[string]IResource
	cacheMutex sync.RWMutex
	eventList  []*event.Event
}

func (m *model) init(db *gorm.DB, resource IResource) error {
	m.db = db
	m.resource = resource
	m.cache = make(map[string]IResource)
	err := m.db.AutoMigrate(resource)
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

func (m *model) getResource(code string) (IResource, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	r, ok := m.cache[code]
	if !ok {
		return nil, false
	}
	return r, true
}

func (m *model) insertResource(r IResource) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.cache[r.GetCode()] = r
}

func (m *model) notify(status string, r IResource, context interface{}) {
	topic := m.resource.Topic()
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

func (m *model) Create(r IResource, context interface{}) (string, error) {
	if m.tx == nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	code := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	r.SetCode(code)
	result := m.getDB().Create(r)
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

func (m *model) Update(r IResource, context interface{}) error {
	if m.tx == nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	temp, err := m.Get(r.GetCode())
	if err != nil {
		printer.Error(err)
		return err
	}
	r.SetID(temp.GetID())
	result := m.getDB().Updates(r)
	if result.Error != nil {
		printer.Error(err)
		return err
	}
	m.deleteResource(r.GetCode())

	temp, err = m.Get(r.GetCode())
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

func (m *model) newResource() IResource {
	t := reflect.TypeOf(m.resource)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface().(IResource)
	}
	return reflect.New(t).Interface().(IResource)
}

func (m *model) Get(code string) (IResource, error) {
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
	slice := reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(m.resource)), 0, 0).Type()).Elem().Interface()
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

