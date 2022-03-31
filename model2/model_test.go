package model2

import (
	"encoding/json"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

type Test struct {
	Data
	Name string `gorm:"column:name;type:VARCHAR(256);not null;default:"`
}

func (t *Test) TableName() string {
	return "t_test"
}

func (t *Test) Topic() string {
	return "topic_test"
}

var m Model

func TestNew(t *testing.T) {
	_, _ = event.Subscribe("topic_test", func(e *event.Event) {
		data, _ := json.Marshal(e)
		printer.Trace(string(data))
	})


	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	m, err = New(db, &Test{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate(t *testing.T) {
	TestModel_GetList(t)

	code, err := m.Create(&Test{
		Data: Data{},
		Name:     "name",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	r, err := m.Get(code)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(r)
	printer.Trace(string(data))

	TestModel_GetList(t)

	err = m.Update(&Test{
		Data: Data{
			Code:      code,
		},
		Name: "update_name",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	r, err = m.Get(code)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(r)
	printer.Trace(string(data))

	TestModel_GetList(t)

	err = m.Delete(code, nil)
	if err != nil {
		t.Fatal(err)
	}
	r, err = m.Get(code)
	if err != nil {
		printer.Error(err)
	}
	data, _ = json.Marshal(r)
	printer.Trace(string(data))

	TestModel_GetList(t)
}

func TestModel_Begin(t *testing.T) {
	m.Begin()
	var err error
	defer func() {
		if err != nil {
			m.Rollback()
		} else {
			m.Commit()
		}
	}()
	var code string
	code, err = m.Create(&Test{
		Data: Data{},
		Name:     "name",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Update(&Test{
		Data: Data{
			Code:      code,
		},
		Name: "update_name",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = m.Delete(code, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestModel_GetList(t *testing.T) {
	values, err := m.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(values)
	printer.Trace(string(data))
}
