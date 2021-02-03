package model

import (
	"encoding/json"
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"sync"
	"testing"
)

var code string
var wg sync.WaitGroup

func TestInit(t *testing.T) {
	subscriber, _ := event.Subscribe("testTopic")
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			e, ok := <-subscriber.Event()
			if !ok {
				break
			}
			switch e.Status {
			case StatusCreated:
				r := e.Data.(*Resource)
				data, _ := json.Marshal(r)
				printer.Trace("Create", string(data))
			case StatusUpdated:
				r := e.Data.(*Resource)
				data, _ := json.Marshal(r)
				printer.Trace("Update", string(data))
			case StatusDeleted:
				r := e.Data.(*Resource)
				data, _ := json.Marshal(r)
				printer.Trace("Delete", string(data))
				subscriber.Unsubscribe()
			}
		}
	}()
}

func TestNew(t *testing.T) {
	db, err := database.NewDatabase("sqlite3", "test.db")
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(db, &Resource{}, "testTopic", true, "t_test")
	if err != nil {
		t.Fatal(err)
	}

	values, err := m.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(values)
	printer.Trace(string(data))

	var r Resource
	r.Name = "123456"
	r.Creator = "5678"
	code, err = m.Create(&r, nil)
	if err != nil {
		t.Fatal(err)
	}

	tempR, err := m.GetByCode(code)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(tempR)
	printer.Trace(string(data))

	r = *(tempR.(*Resource))
	r.Code = code
	r.Name = "ceshi1"
	err = m.UpdateByCode(&r, nil)
	if err != nil {
		t.Fatal(err)
	}

	tempR, err = m.GetByCode(code)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(tempR)
	printer.Trace(string(data))

	err = m.DeleteByCode(code, nil)
	if err != nil {
		t.Fatal(err)
	}
	values, err = m.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(values)
	printer.Trace(string(data))

	tempR, err = m.GetByCode(code)
	if err != nil {
		printer.Error(err)
	}
	wg.Wait()
}
