package notifier

import (
	"encoding/json"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/controller"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/server"
	"sync"
	"time"
)

type FilterFunc func(session *controller.Session, e *event.Event) bool

func SetFilter(topic string, filterFunc FilterFunc) {
	notifier.filterMutex.Lock()
	defer notifier.filterMutex.Unlock()
	notifier.filterMap[topic] = filterFunc
}

var notifier *notifierHandler

func init() {
	notifier = new(notifierHandler)
	notifier.socketMap = make(map[string]server.Websocket)
	notifier.filterMap = make(map[string]FilterFunc)
	age := config.GetInt("session.age") / 2
	if age == 0 {
		age = controller.DefaultAge / 2
	}
	notifier.timer = time.NewTimer(time.Duration(age) * time.Second)
	go notifier.updateSessionLoop()
	notifier.subscriber, _ = event.SubscribeAll()
	go notifier.eventLoop()
	server.Router().GET("/ws/notification", controller.SessionMiddle(), server.GenerateWebsocketHandlerFunc(notifier))
}

type notifierHandler struct {
	socketMap   map[string]server.Websocket
	socketMutex sync.Mutex
	timer       *time.Timer
	subscriber  event.Subscriber
	filterMap   map[string]FilterFunc
	filterMutex sync.Mutex
}

func (n *notifierHandler) NewConnection(socket server.Websocket) {
	token, _ := socket.Context().Get("Token")
	_, err := controller.GetSession(token.(string))
	if err != nil {
		printer.Error(err)
		_ = socket.Close()
		return
	}
	n.socketMutex.Lock()
	defer n.socketMutex.Unlock()

	_, ok := n.socketMap[token.(string)]
	if ok {
		printer.Error("Token is conflicted. ")
		_ = socket.Close()
		return
	}

	n.socketMap[token.(string)] = socket
}

func (n *notifierHandler) Disconnected(socket server.Websocket) {
	token, _ := socket.Context().Get("Token")
	n.socketMutex.Lock()
	defer n.socketMutex.Unlock()
	delete(n.socketMap, token.(string))
}

func (n *notifierHandler) ReadMessage(socket server.Websocket, message []byte) {

}

func (n *notifierHandler) ReadBytes(socket server.Websocket, bytes []byte) {

}

func (n *notifierHandler) updateSessionLoop() {
	for {
		select {
		case <-n.timer.C:
			tokenList := n.getTokenList()
			for _, token := range tokenList {
				controller.UpdateSession(token)
			}
		}
		age := config.GetInt("session.age") / 2
		if age == 0 {
			age = controller.DefaultAge / 2
		}
		n.timer.Reset(time.Duration(age) * time.Second)
	}
}

func (n *notifierHandler) getWebsocket(token string) (ws server.Websocket, ok bool) {
	n.socketMutex.Lock()
	defer n.socketMutex.Unlock()
	ws, ok = n.socketMap[token]
	return
}

func (n *notifierHandler) getTokenList() (tokenList []string) {
	n.socketMutex.Lock()
	defer n.socketMutex.Unlock()
	for token, _ := range n.socketMap {
		tokenList = append(tokenList, token)
	}
	return
}

func (n *notifierHandler) getFilter(topic string) (filter FilterFunc, ok bool) {
	n.filterMutex.Lock()
	defer n.filterMutex.Unlock()
	filter, ok = n.filterMap[topic]
	return
}

func (n *notifierHandler) eventLoop() {
	for {
		e, ok := <-n.subscriber.Event()
		if !ok {
			printer.Error("quit...")
			break
		}
		switch e.Topic {
		case controller.TopicSession:
			if e.Status == model.StatusDeleted {
				session, ok := e.Data.(*controller.Session)
				if !ok {
					continue
				}
				ws, ok := n.getWebsocket(session.Token)
				if !ok {
					continue
				}
				_ = ws.Close()
			}
		default:
			tempEvent := *e
			tempEvent.Context = nil
			data, _ := json.Marshal(tempEvent)

			tokenList := n.getTokenList()
			for _, token := range tokenList {
				filter, ok := n.getFilter(e.Topic)
				if ok {
					session, err := controller.GetSession(token)
					if err != nil {
						ws, ok := n.getWebsocket(token)
						if ok {
							_ = ws.Close()
						}
						continue
					}
					if filter(session, e) == false {
						continue
					}
				}
				ws, ok := n.getWebsocket(token)
				if !ok {
					continue
				}
				if err := ws.WriteMessage(data); err != nil {
					printer.Error(err)
				}
			}
		}
	}
	n.subscriber.Unsubscribe()
	n.subscriber, _ = event.SubscribeAll()
	go n.eventLoop()
}
