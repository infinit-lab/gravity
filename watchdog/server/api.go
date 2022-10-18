package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/infinit-lab/gravity/controller"
	"github.com/infinit-lab/gravity/printer"
	"time"
)

var c controller.Controller

const (
	StatusRunning string = "running"
	StatusExit    string = "exit"
	StatusStop    string = "stop"

	OperationStart            string = "start"
	OperationStop             string = "stop"
	OperationRestart          string = "restart"
	OperationAutoStart        string = "autoStart"
	OperationDisableAutoStart string = "disableAutoStart"
)

type Status struct {
	Name            string `json:"name"`
	Status          string `json:"status"`
	PId             int    `json:"pid"`
	RunningDuration string `json:"runningDuration"`
	LastError       string `json:"lastError"`
}

func initController() {
	c, _ = controller.New("/api/watchdog")
	c.GET("/status", func(context *gin.Context, session *controller.Session) (i interface{}, e error) {
		processList, _ := getProcessList()
		statusList := make([]*Status, 0, len(processList))
		for _, process := range processList {
			s, err := generateStatus(process.Id)
			if err != nil {
				printer.Error(err)
				continue
			}
			s.Name = process.Name
			statusList = append(statusList, s)
		}
		return statusList, nil
	})
	c.GET("/status/:name", func(context *gin.Context, session *controller.Session) (i interface{}, e error) {
		name := context.Param("name")
		process, err := getProcessByName(name)
		if err != nil {
			return nil, err
		}
		s, err := generateStatus(process.Id)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		s.Name = process.Name
		return []*Status{s}, nil
	})
	c.POST("/process", func(context *gin.Context, session *controller.Session) (i interface{}, e error) {
		var p process
		err := controller.GetRequestBody(context, &p)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		return nil, createProcess(&p)
	})
	c.DELETE("/process/:name", func(context *gin.Context, session *controller.Session) (i interface{}, e error) {
		return nil, deleteProcess(context.Param("name"))
	})
	c.PUT("/process/:name/:operation", func(context *gin.Context, session *controller.Session) (i interface{}, e error) {
		operation := context.Param("operation")
		name := context.Param("name")
		switch operation {
		case OperationStart:
			return nil, start(name)
		case OperationStop:
			return nil, stop(name)
		case OperationRestart:
			err := stop(name)
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			return nil, start(name)
		case OperationAutoStart:
			p, err := getProcessByName(name)
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			t := *p
			t.IsAutoStart = true
			return nil, updateProcess(&t)
		case OperationDisableAutoStart:
			p, err := getProcessByName(name)
			if err != nil {
				printer.Error(err)
				return nil, err
			}
			t := *p
			t.IsAutoStart = false
			return nil, updateProcess(&t)
		}
		return nil, nil
	})
}

func generateStatus(id int) (*Status, error) {
	s := getStatus(id)
	if s == nil {
		return nil, errors.New(fmt.Sprintf("%d not exists. ", id))
	}
	t := new(Status)
	if !s.isStarted {
		t.Status = StatusStop
		t.PId = 0
		t.RunningDuration = "00:00:00"
		t.LastError = ""
	} else {
		if s.isRunning {
			t.Status = StatusRunning
			if s.command != nil {
				t.PId = s.command.Process.Pid
			}
			t.RunningDuration = time.Now().Sub(s.startTime).String()
			t.LastError = ""
		} else {
			t.Status = StatusExit
			t.PId = 0
			t.RunningDuration = "00:00:00"
			t.LastError = s.lastError
		}
	}
	return t, nil
}
