package main

import (
	"errors"
	"fmt"
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type process struct {
	model.Resource
	AbstractPath string `json:"abstractPath" db:"abstractPath" db_type:"VARCHAR(256)" db_default:"''"`
	CommandLine  string `json:"commandLine" db:"commandLine" db_type:"VARCHAR(1024)" db_default:"''"`
	IsAutoStart  bool   `json:"isAutoStart" db:"isAutoStart" db_default:"0"`
}

var processDatabase database.Database
var processModel model.Model

const topicProcess = "model.process"

func initModel() {
	var err error
	processDatabase, err = database.NewDatabase("sqlite3", "./process.db")
	if err != nil {
		printer.Error(err)
		panic(err)
	}
	processModel, err = model.New(processDatabase, &process{}, topicProcess, true, "t_process")
	if err != nil {
		printer.Error(err)
		panic(err)
	}
}

func getProcessByName(name string) (*process, error) {
	values, err := processModel.GetListWithWhereSql("WHERE `name` = ?", name)
	if err != nil {
		printer.Error(err)
		return nil, errors.New(fmt.Sprintf("%s not exists. ", name))
	}
	if len(values) == 0 {
		return nil, errors.New(fmt.Sprintf("%s not exists. ", name))
	}
	return values[0].(*process), nil
}

func createProcess(p *process) error {
	if len(p.Name) == 0 || len(p.AbstractPath) == 0 {
		return errors.New("Name or abstract path is empty. ")
	}
	_, err := getProcessByName(p.Name)
	if err == nil {
		return errors.New(fmt.Sprintf("%s exists. ", p.Name))
	}
	_, err = processModel.Create(p, nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func updateProcess(p *process) error {
	return processModel.Update(p, nil)
}

func deleteProcess(name string) error {
	p, err := getProcessByName(name)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = processModel.Delete(p.Id, nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func getProcessList() ([]*process, error) {
	values, err := processModel.GetList()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	processList := make([]*process, len(values))
	for i, value := range values {
		processList[i] = value.(*process)
	}
	return processList, nil
}
