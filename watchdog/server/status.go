package main

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

type status struct {
	isStarted bool
	isRunning bool
	command   *exec.Cmd
	cancel   context.CancelFunc
	lastError string
	startTime time.Time
}

var statusMap = make(map[int]*status)
var statusMutex sync.Mutex

func isNeedRun(id int) bool {
	var s *status
	statusMutex.Lock()
	s = statusMap[id]
	statusMutex.Unlock()
	if s == nil {
		return false
	}
	return s.isRunning == false && s.isStarted
}

func isNeedStop(id int) bool {
	var s *status
	statusMutex.Lock()
	s = statusMap[id]
	statusMutex.Unlock()
	if s == nil {
		return false
	}
	return s.isStarted == false && s.isRunning
}

func started(id int) {
	statusMutex.Lock()
	s, ok := statusMap[id]
	if !ok {
		s = &status{
			isStarted: true,
		}
		statusMap[id] = s
	}
	s.isStarted = true
	statusMutex.Unlock()
}

func stopped(id int) {
	statusMutex.Lock()
	s, ok := statusMap[id]
	if !ok {
		s = &status{
			isStarted: false,
			isRunning: false,
			command:   nil,
		}
		statusMap[id] = s
	}
	s.isStarted = false
	statusMutex.Unlock()
}

func running(id int, c *exec.Cmd, cancel context.CancelFunc) {
	statusMutex.Lock()
	s, ok := statusMap[id]
	if ok {
		s.isRunning = true
		s.command = c
		s.cancel = cancel
		s.lastError = ""
		s.startTime = time.Now()
	}
	statusMutex.Unlock()
}

func quit(id int, e string) {
	statusMutex.Lock()
	s, ok := statusMap[id]
	if ok {
		s.isRunning = false
		s.command = nil
		s.cancel = nil
		s.lastError = e
	}
	statusMutex.Unlock()
}

func deleteStatus(id int) {
	statusMutex.Lock()
	delete(statusMap, id)
	statusMutex.Unlock()
}

func getCommandAndCancelFunc(id int) (cmd *exec.Cmd, cancel context.CancelFunc) {
	statusMutex.Lock()
	s, ok := statusMap[id]
	if ok {
		cmd = s.command
		cancel = s.cancel
	}
	statusMutex.Unlock()
	return
}

func getStatus(id int) (s *status) {
	statusMutex.Lock()
	s = statusMap[id]
	statusMutex.Unlock()
	return
}
