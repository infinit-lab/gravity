package database

import (
	"encoding/json"
	"github.com/infinit-lab/gravity/printer"
	"testing"
)

var db Database
var tb Table

type TestResource struct {
	Resource
	Uuid string `json:"uuid" db:"uuid" db_index:"index" db_type:"VARCHAR(32)" db_default:"''"`
	Time string `json:"time" db:"time" db_index:"index" db_type:"DATETIME" db_omit:"create,update" db_default:"CURRENT_TIMESTAMP"`
}

func TestDatabase(t *testing.T) {
	var err error
	db, err = NewDatabase("sqlite3", "test.db")
	if err != nil {
		t.Fatal(err)
	}
	tb, err = db.NewTable(TestResource{}, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTable_Create(t *testing.T) {
	r := TestResource{
		Resource: Resource{
			Name:    "TestName",
			Creator: "TestCreator",
			Updater: "TestUpdater",
		},
		Uuid: "12345",
	}

	// Create
	ret, err := tb.Create(&r)
	if err != nil {
		t.Fatal(err)
	}

	values, err := tb.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(values)
	printer.Trace(string(data))

	ret, err = tb.Create(&r)
	if err != nil {
		t.Fatal(err)
	}

	values, err = tb.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(values)
	printer.Trace(string(data))

	id, err := ret.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	value, err := tb.Get("WHERE `id` = ? LIMIT 1", id)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(value)
	printer.Trace(string(data))

	// Update
	r.Name = "TestUpdateName"
	ret, err = tb.Update(r, "WHERE `id` = ?", id)
	if err != nil {
		t.Fatal(err)
	}
	values, err = tb.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(values)
	printer.Trace(string(data))

	value, err = tb.Get("WHERE `id` = ? LIMIT 1", id)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(value)
	printer.Trace(string(data))

	// Delete
	ret, err = tb.Delete("WHERE `id` = ?", id)
	if err != nil {
		t.Fatal(err)
	}
	values, err = tb.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(values)
	printer.Trace(string(data))

	// Delete All
	ret, err = tb.Delete("")
	if err != nil {
		t.Fatal(err)
	}
	values, err = tb.GetList("")
	if err != nil {
		t.Fatal(err)
	}
	data, _ = json.Marshal(values)
	printer.Trace(string(data))
}

func TestDatabase_Close(t *testing.T) {
	db.Close()
}
