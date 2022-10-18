package reset

import (
	"github.com/infinit-lab/gravity/printer"
	"testing"
)

type testReseter struct {
	name string
	depends []string
}

func (r *testReseter)Name() string {
	return r.name
}

func (r *testReseter)Depends() []string {
	return r.depends
}

func (r *testReseter)Check(data interface{}) (interface{}, error) {
	printer.Trace(r.name, "check")
	return nil, nil
}

func (r *testReseter)Reset(data interface{}, checkData interface{}) (interface{}, bool, error) {
	printer.Trace(r.name, "reset")
	return nil, false, nil
}

func (r *testReseter)AfterReset(data interface{}, checkData interface{}, resetData interface{}, isExit bool) error {
	printer.Trace(r.name, "after")
	return nil
}

func TestRegisterReseter(t *testing.T) {
	r := new(testReseter)
	r.name = "1"
	r.depends = nil
	err := RegisterReseter(r)
	if err != nil {
		printer.Error(err)
		return
	}
	r = new(testReseter)
	r.name = "2"
	r.depends = append(r.depends, "1")
	err = RegisterReseter(r)
	if err != nil {
		printer.Error(err)
		return
	}
	r = new(testReseter)
	r.name = "3"
	r.depends = append(r.depends, "2")
	err = RegisterReseter(r)
	if err != nil {
		printer.Error(err)
		return
	}
	r = new(testReseter)
	r.name = "4"
	r.depends = append(r.depends, "1")
	err = RegisterReseter(r)
	if err != nil {
		printer.Error(err)
		return
	}
	err = RegisterReseter(new(testReseter))
	if err != nil {
		printer.Error(err)
	}
	r = new(testReseter)
	r.name = "1"
	err = RegisterReseter(r)
	if err != nil {
		printer.Error(err)
	}
	r = new(testReseter)
	r.name = "5"
	r.depends = append(r.depends, "5")
	err = RegisterReseter(r)
	if err != nil {
		printer.Error(err)
	}
}

func TestReset(t *testing.T) {
	err := Reset(nil)
	if err != nil {
		printer.Error(err)
	}
}
