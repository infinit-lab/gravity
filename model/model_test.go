package model

import (
	"encoding/json"
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"sync"
	"testing"
	"time"
)

const (
	TopicModelTest = "model_test"
)

type TestResource struct {
	Resource
	Description string `json:"description" db:"description"`
}

func TestNewModel(t *testing.T) {
	subscriber, err := event.Subscribe("model_test")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			e, ok := <-subscriber.Event()
			if !ok {
				break
			}
			data, _ := json.Marshal(e)
			printer.Trace(string(data))
		}
	}()

	db, err := database.NewDatabase("sqlite3", "test.db")
	if err != nil {
		t.Fatal(err)
	}

	m, err := New(db, &TestResource{}, TopicModelTest, true, "")
	if err != nil {
		t.Fatal(err)
	}

	var r TestResource
	r.Name = "Create"
	r.Creator = "TestCreator"
	r.Description = "CreateDescription"

	id, err := m.Create(&r, nil)
	if err != nil {
		t.Fatal(err)
	}

	value, err := m.Get(id)
	if err != nil {
		t.Fatal()
	}
	data, _ := json.Marshal(value)
	printer.Trace(string(data))

	r.Id = id
	r.Name = "Update"
	r.Updater = "TestUpdater"
	r.Description = "UpdateDescription"
	r.UpdateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	err = m.Update(&r, nil)
	if err != nil {
		t.Fatal(err)
	}

	value, err = m.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(value)
	printer.Trace(string(data))

	values, err := m.GetList()
	if err != nil {
		t.Fatal()
	}
	data, _ = json.Marshal(values)
	printer.Trace(string(data))

	err = m.Delete(id, nil)
	if err != nil {
		t.Fatal(err)
	}

	values, err = m.GetList()
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(values)
	printer.Trace(string(data))

	db.Close()
	subscriber.Unsubscribe()
	wg.Wait()
}
