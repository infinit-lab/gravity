package main

import (
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/server"
	"sync"
)

type websocketHandler struct {
	server.WebsocketHandler
	subscribers map[string]map[server.Websocket]bool
	subscribersMutex sync.Mutex
}

var h = new(websocketHandler)

func initLog() {
	h.subscribers = make(map[string]map[server.Websocket]bool)
	server.Router().GET("/ws/log/:name", server.GenerateWebsocketHandlerFunc(h))
	
	_, _ = event.Subscribe(topicProcess, func(e *event.Event) {
		if e.Status != model.StatusDeleted {
			return
		}
		p := e.Data.(*process)
		h.subscribersMutex.Lock()
		subscribes, ok := h.subscribers[p.Name]
		if ok {
			for socket, _ := range subscribes {
				_ = socket.Close()
			}
			delete(h.subscribers, p.Name)
		}
		h.subscribersMutex.Unlock()
	})
}

func (h *websocketHandler) NewConnection(socket server.Websocket) {
	name := socket.Context().Param("name")
	_, err := getProcessByName(name)
	if err != nil {
		_ = socket.Close()
		return
	}
	h.subscribersMutex.Lock()
	subscribers, ok := h.subscribers[name]
	if !ok {
	 	subscribers = make(map[server.Websocket]bool)
		h.subscribers[name] = subscribers
	}
	subscribers[socket] = true
	h.subscribersMutex.Unlock()
}

func (h *websocketHandler) Disconnected(socket server.Websocket) {
	h.subscribersMutex.Lock()
	name := socket.Context().Param("name")
	subscribers, ok := h.subscribers[name]
	if !ok {
		subscribers = make(map[server.Websocket]bool)
		h.subscribers[name] = subscribers
	}
	delete(subscribers, socket)
	h.subscribersMutex.Unlock()
}
func (h *websocketHandler) ReadMessage(socket server.Websocket, message []byte) {}
func (h *websocketHandler) ReadBytes(socket server.Websocket, bytes []byte) {}

func remoteLog(name string, log []byte) {
	h.subscribersMutex.Lock()
	subscribers, ok := h.subscribers[name]
	if ok {
		for socket, _ := range subscribers {
			_ = socket.WriteMessage(log)
		}
	}
	h.subscribersMutex.Unlock()
}
