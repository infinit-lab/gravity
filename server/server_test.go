package server

import (
	"device_framework/config"
	"device_framework/printer"
	"github.com/gorilla/websocket"
	"net/url"
	"os"
	"sync"
	"testing"
)

type websocketHandler struct {
}

func (w *websocketHandler) NewConnection(socket Websocket) {
	printer.Trace("NewConnection")
}

func (w *websocketHandler) Disconnected(socket Websocket) {
	printer.Trace("Disconnected")
}

func (w *websocketHandler) ReadMessage(socket Websocket, message []byte) {
	printer.Trace(string(message))
	if err := socket.WriteMessage(message); err != nil {
		printer.Error(err)
	}
}

func (w *websocketHandler) ReadBytes(socket Websocket, bytes []byte) {
	printer.Trace(string(bytes))
	if err := socket.WriteBytes(bytes); err != nil {
		printer.Error(err)
	}
}

func TestWebsocket(t *testing.T) {
	handler := new(websocketHandler)
	Router().GET("/ws", GenerateWebsocketHandlerFunc(handler))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		os.Args = append(os.Args, "server.port=8081")
		config.LoadArgs()
		if err := Run(); err != nil {
			printer.Error(err)
		}
	}()

	u := url.URL{
		Scheme: "ws",
		Host:   "127.0.0.1:8081",
		Path:   "/ws",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.WriteMessage(websocket.TextMessage, []byte("123456"))
	msgType, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	printer.Trace(msgType)
	printer.Trace(string(message))

	_ = conn.WriteMessage(websocket.BinaryMessage, []byte("567890"))
	msgType, message, err = conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	printer.Trace(msgType)
	printer.Trace(string(message))

	_ = conn.Close()

	if err := Shutdown(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
