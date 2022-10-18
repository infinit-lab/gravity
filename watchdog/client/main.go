package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/version"
	"io/ioutil"
	"net/http"
)

type Command string

const (
	CommandCreate  Command = "create"
	CommandDelete  Command = "delete"
	CommandStatus  Command = "status"
	CommandStart   Command = "start"
	CommandStop    Command = "stop"
	CommandRestart Command = "restart"
	CommandLog     Command = "log"
	CommandEnable  Command = "enable"
	CommandDisable Command = "disable"

	Port int = 10000
)

func main() {
	helpFlag := flag.Bool("h", false, "打印帮助")
	if *helpFlag {
		help()
		return
	}
	host := flag.String("H", "127.0.0.1", "主机地址")
	commandLine := flag.String("c", "", "命令行参数")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		help()
		return
	}
	command := Command(args[0])
	name := ""
	if len(args) > 1 {
		name = args[1]
	}
	path := ""
	if len(args) > 2 {
		path = args[2]
	}
	switch command {
	case CommandCreate:
		createProcess(*host, name, path, *commandLine)
	case CommandDelete:
		deleteProcess(*host, name)
	case CommandStatus:
		status(*host, name)
	case CommandStart:
		start(*host, name)
	case CommandStop:
		stop(*host, name)
	case CommandRestart:
		restart(*host, name)
	case CommandLog:
		log(*host, name)
	case CommandEnable:
		enable(*host, name)
	case CommandDisable:
		disable(*host, name)
	default:
		help()
	}
}

func help() {
	fmt.Println("watchdogctl 版本", version.GetVersion())
	fmt.Println()
	fmt.Println("watchdogctl [-H 主机地址] [-c 命令行传参] create 进程名称 进程路径")
	fmt.Println("watchdogctl [-H 主机地址] delete  进程名称  ")
	fmt.Println("watchdogctl [-H 主机地址] status  [进程名称]")
	fmt.Println("watchdogctl [-H 主机地址] start   进程名称  ")
	fmt.Println("watchdogctl [-H 主机地址] stop    进程名称  ")
	fmt.Println("watchdogctl [-H 主机地址] restart 进程名称  ")
	fmt.Println("watchdogctl [-H 主机地址] log     进程名称  ")
	fmt.Println("watchdogctl [-H 主机地址] enable  进程名称  ")
	fmt.Println("watchdogctl [-H 主机地址] disable 进程名称  ")
}

type response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func request(host string, method string, uri string, body []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s:%d%s", host, Port, uri), bytes.NewReader(body))
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	rsp, err := client.Do(req)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	data, err := ioutil.ReadAll(rsp.Body)
	_ = rsp.Body.Close()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var resp response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New(resp.Message)
	}
	data, _ = json.Marshal(resp.Data)
	return data, nil
}

type process struct {
	Name         string `json:"name"`
	AbstractPath string `json:"abstractPath"`
	CommandLine  string `json:"commandLine"`
	IsAutoStart  bool   `json:"isAutoStart"`
}

func createProcess(host string, name string, path string, commandLine string) {
	p := process{
		Name:         name,
		AbstractPath: path,
		CommandLine:  commandLine,
		IsAutoStart:  true,
	}
	data, err := json.Marshal(p)
	if err != nil {
		printer.Error(err)
		return
	}
	_, err = request(host, http.MethodPost, "/api/watchdog/process", data)
	if err != nil {
		printer.Error(err)
		return
	}
}

func deleteProcess(host string, name string) {
	_, err := request(host, http.MethodDelete, fmt.Sprintf("/api/watchdog/process/%s", name), []byte{})
	if err != nil {
		printer.Error(err)
		return
	}
}

type processStatus struct {
	Name            string `json:"name"`
	Pid             int    `json:"pid"`
	Status          string `json:"status"`
	RunningDuration string `json:"runningDuration"`
	LastError       string `json:"lastError"`
}

func status(host string, name string) {
	var data []byte
	var err error
	if len(name) == 0 {
		data, err = request(host, http.MethodGet, "/api/watchdog/status", []byte{})
	} else {
		data, err = request(host, http.MethodGet, fmt.Sprintf("/api/watchdog/status/%s", name), []byte{})
	}
	if err != nil {
		printer.Error(err)
		return
	}
	var statusList []*processStatus
	err = json.Unmarshal(data, &statusList)
	if err != nil {
		printer.Error(err)
		return
	}
	fmt.Println(fmt.Sprintf("%-10s %-6s %-10s %-25s %-s", "NAME", "PID", "STATUS", "TIME", "ERROR"))
	for _, status := range statusList {
		fmt.Println(fmt.Sprintf("%-10s %-6d %-10s %-25s %-s", status.Name, status.Pid, status.Status, status.RunningDuration, status.LastError))
	}
}

func start(host string, name string) {
	_, err := request(host, http.MethodPut, fmt.Sprintf("/api/watchdog/process/%s/start", name), []byte{})
	if err != nil {
		printer.Error(err)
		return
	}
}

func stop(host string, name string) {
	_, err := request(host, http.MethodPut, fmt.Sprintf("/api/watchdog/process/%s/stop", name), []byte{})
	if err != nil {
		printer.Error(err)
		return
	}
}

func restart(host string, name string) {
	_, err := request(host, http.MethodPut, fmt.Sprintf("/api/watchdog/process/%s/restart", name), []byte{})
	if err != nil {
		printer.Error(err)
		return
	}
}

func log(host string, name string) {
	if len(name) == 0 {
		return
	}
	ws, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%d/ws/log/%s", host, Port, name), nil)
	if err != nil {
		printer.Error(err)
		return
	}
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			printer.Error(err)
			break
		}
		fmt.Print(string(data))
	}
	_ = ws.Close()
}

func enable(host string, name string) {
	_, err := request(host, http.MethodPut, fmt.Sprintf("/api/watchdog/process/%s/autoStart", name), []byte{})
	if err != nil {
		printer.Error(err)
		return
	}
}

func disable(host string, name string) {
	_, err := request(host, http.MethodPut, fmt.Sprintf("/api/watchdog/process/%s/disableAutoStart", name), []byte{})
	if err != nil {
		printer.Error(err)
		return
	}
}
