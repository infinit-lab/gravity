package module

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/controller"
	"github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/sqlite"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var DefaultDatabase sqlite.Database

type Module interface {
	Controller() controller.Controller
	Model() model.Model
}

func New(prefix string, topic string, resource model.Id, omitMethod ...string) (Module, error) {
	c, err := controller.New(prefix)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	m, err := model.New(DefaultDatabase, resource, topic, true)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	module := new(module)
	module.c = c
	module.m = m
	module.resource = resource

	isGet := true
	isPost := true
	isPut := true
	isDelete := true

	for _, m := range omitMethod {
		switch m {
		case http.MethodGet:
			isGet = false
		case http.MethodPost:
			isPost = false
		case http.MethodPut:
			isPut = false
		case http.MethodDelete:
			isDelete = false
		}
	}

	if isGet {
		c.GETWithSession("/data", func(context *gin.Context, session *controller.Session) (interface{}, error) {
			if session == nil {
				return nil, errors.New("Unauthorized. ")
			}
			return m.GetList()
		})
		c.GETWithSession("/data/:id", func(context *gin.Context, session *controller.Session) (interface{}, error) {
			if session == nil {
				return nil, errors.New("Unauthorized. ")
			}
			id, err := strconv.Atoi(context.Param("id"))
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			return m.Get(id)
		})
	}

	if isPost {
		c.POSTWithSession("/data", func(context *gin.Context, session *controller.Session) (interface{}, error) {
			if session == nil {
				return nil, errors.New("Unauthorized. ")
			}
			data, err := context.GetRawData()
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			r := module.NewResource()
			err = json.Unmarshal(data, r)

			temp, ok := module.ConvertToResource(r)
			if ok {
				temp.Creator = session.Username
			}
			_, err = m.Create(r, session)
			return nil, err
		})
	}

	if isPut {
		c.PUTWithSession("/data/:id", func(context *gin.Context, session *controller.Session) (interface{}, error) {
			if session == nil {
				return nil, errors.New("Unauthorized. ")
			}
			id, err := strconv.Atoi(context.Param("id"))
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			data, err := context.GetRawData()
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			r := module.NewResource()
			err = json.Unmarshal(data, r)
			r.SetId(id)

			temp, ok := module.ConvertToResource(r)
			if ok {
				temp.Updater = session.Username
				temp.UpdateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
			}
			err = m.Update(r, session)
			return nil, err
		})
	}

	if isDelete {
		c.DELETEWithSession("/data/:id", func(context *gin.Context, session *controller.Session) (interface{}, error) {
			if session == nil {
				return nil, errors.New("Unauthorized. ")
			}
			id, err := strconv.Atoi(context.Param("id"))
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			err = m.Delete(id, session)
			return nil, err
		})
	}
	c.PUT("/cache/sync/", func(context *gin.Context, session *controller.Session) (interface{}, error) {
		err := m.Sync()
		return nil, err
	})
	return module, nil
}

func init() {
	database := config.GetString("sqlite.database")
	if len(database) == 0 {
		database = "database.db"
	}
	var err error
	DefaultDatabase, err = sqlite.NewDatabase(database)
	if err != nil {
		DefaultDatabase = nil
		printer.Error(err)
	}
}

type module struct {
	c        controller.Controller
	m        model.Model
	resource model.Id
}

func (m *module) Controller() controller.Controller {
	return m.c
}

func (m *module) Model() model.Model {
	return m.m
}

func (m *module) NewResource() model.Id {
	t := reflect.TypeOf(m.resource)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface().(model.Id)
}

func (m *module) ConvertToResource(r interface{}) (*model.Resource, bool) {
	v := reflect.ValueOf(r)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.CanAddr() {
		resource, ok := v.Addr().Interface().(*model.Resource)
		if ok {
			return resource, ok
		}
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr {
			if field.CanInterface() {
				resource, ok := m.ConvertToResource(field.Interface())
				if ok {
					return resource, ok
				}
			}
		} else if field.Kind() == reflect.Struct {
			if field.CanAddr() {
				resource, ok := m.ConvertToResource(field.Addr().Interface())
				if ok {
					return resource, ok
				}
			}
		}
	}
	return nil, false
}
