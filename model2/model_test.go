package model2

import (
	"encoding/json"
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	mdl "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
	"sync"
	"testing"
	"time"
)

var id int
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
				r := e.Data.(*mdl.Resource)
				id = r.Id
				data, _ := json.Marshal(r)
				printer.Trace("Create", string(data))
			case StatusUpdated:
				r := e.Data.(*mdl.Resource)
				id = r.Id
				data, _ := json.Marshal(r)
				printer.Trace("Update", string(data))
			case StatusDeleted:
				r := e.Data.(*mdl.Resource)
				id = r.Id
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
	m, err := New(db, &mdl.Resource{}, "testTopic", true, "t_test")
	if err != nil {
		t.Fatal(err)
	}

	values, err := m.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(values)
	printer.Trace(string(data))

	var r mdl.Resource
	r.Name = "123456"
	r.Creator = "5678"
	err = m.Create(&r, nil)
	if err != nil {
		t.Fatal(err)
	}

	for id == 0 {
		time.Sleep(100 * time.Millisecond)
	}
	tempR, err := m.Get("WHERE `id` = ?", id)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(tempR)
	printer.Trace(string(data))

	r = *(tempR.(*mdl.Resource))
	r.Name = "ceshi1"
	err = m.Update(r, nil, "WHERE `id` = ?", r.Id)
	if err != nil {
		t.Fatal(err)
	}

	tempR, err = m.Get("WHERE `id` = ?", id)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(tempR)
	printer.Trace(string(data))

	err = m.Delete(nil, "WHERE `id` = ?", r.Id)
	if err != nil {
		t.Fatal(err)
	}
	tempR, err = m.Get("WHERE `id` = ?", id)
	if err != nil {
		printer.Error(err)
	}
	wg.Wait()
}
