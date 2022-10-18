package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
	"io"
	"os/exec"
	"path"
	"strings"
	"time"
)

var cmdTicker = time.NewTicker(time.Second)

func initCmd() {
	processList, _ := getProcessList()
	for _, process := range processList {
		if process.IsAutoStart {
			started(process.Id)
		} else {
			stopped(process.Id)
		}
	}

	_, _ = event.Subscribe(topicProcess, func(e *event.Event) {
		p := e.Data.(*process)
		switch e.Status {
		case model.StatusCreated:
			started(p.Id)
		case model.StatusDeleted:
			_ = stop(p.Name)
			deleteStatus(p.Id)
		}
	})

	go func() {
		for {
			<-cmdTicker.C
			processList, _ := getProcessList()
			for _, process := range processList {
				if isNeedStop(process.Id) {
					_, cancelFunc := getCommandAndCancelFunc(process.Id)
					if cancelFunc != nil {
						cancelFunc()
					}
					continue
				}
				if !isNeedRun(process.Id) {
					continue
				}
				ctx, cancelFunc := context.WithCancel(context.Background())
				commandLines := strings.Split(process.CommandLine, " ")
				cmd := exec.CommandContext(ctx, process.AbstractPath, commandLines...)
				cmd.Dir = path.Dir(process.AbstractPath)
				outputPipe, err := cmd.StdoutPipe()
				if err != nil {
					printer.Error(err)
				}

				if err := cmd.Start(); err != nil {
					printer.Error(err)
					quit(process.Id, err.Error())
					if outputPipe != nil {
						_ = outputPipe.Close()
					}
					continue
				}

				if outputPipe != nil {
					go func(name string, outputPipe io.ReadCloser) {
						buffer := make([]byte, 1024)
						for {
							n, err := outputPipe.Read(buffer)
							if err != nil {
								break
							}
							remoteLog(name, buffer[:n])
						}
						_ = outputPipe.Close()
					}(process.Name, outputPipe)
				}

				go func(id int, cmd *exec.Cmd, ctx context.Context) {
					// <- ctx.Done()
					// printer.Tracef("Process %d quit. ", cmd.Process.Pid)
					_ = cmd.Wait()
					printer.Tracef("Process %d quit. ", cmd.Process.Pid)
					quit(id, "")
				}(process.Id, cmd, ctx)

				running(process.Id, cmd, cancelFunc)
			}
		}
	}()
}

func start(name string) error {
	process, err := getProcessByName(name)
	if err != nil {
		return errors.New(fmt.Sprintf("%s not exists. ", name))
	}
	started(process.Id)
	return nil
}

func stop(name string) error {
	process, err := getProcessByName(name)
	if err != nil {
		return errors.New(fmt.Sprintf("%s not exists. ", name))
	}
	stopped(process.Id)
	return nil
}
