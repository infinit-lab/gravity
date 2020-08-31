package event

import (
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"sync"
)

type Event struct {
	Topic   string      `json:"topic"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Context interface{} `json:"context,omitempty"`
}

type Subscriber interface {
	Event() <-chan *Event
	Unsubscribe()
}

func Subscribe(topic string) (Subscriber, error) {
	s := new(subscriber)
	s.c = make(chan *Event)
	s.topic = topic

	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	list := subscriberMap[topic]
	list = append(list, s)
	subscriberMap[topic] = list
	return s, nil
}

func SubscribeAll() (Subscriber, error) {
	return Subscribe("")
}

func Publish(event *Event) error {
	if event == nil {
		return errors.New("The event is nil.")
	}

	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	publish(event.Topic, event)
	publish("", event)
	return nil
}

var subscriberMap map[string][]Subscriber
var subscriberMutex sync.Mutex

func init() {
	subscriberMap = make(map[string][]Subscriber)
}

type subscriber struct {
	c     chan *Event
	topic string
}

func (s *subscriber) Event() <-chan *Event {
	return s.c
}

func (s *subscriber) Unsubscribe() {
	subscriberMutex.Lock()
	defer func() {
		close(s.c)
		subscriberMutex.Unlock()
	}()

	subscriberList, ok := subscriberMap[s.topic]
	if !ok {
		return
	}
	for i, subscriber := range subscriberList {
		if subscriber == s {
			var list []Subscriber
			list = append(list, subscriberList[0:i]...)
			if i+1 < len(subscriberList) {
				list = append(list, subscriberList[i+1:]...)
			}
			subscriberMap[s.topic] = list
			break
		}
	}
}

func publish(topic string, event *Event) {
	list, ok := subscriberMap[topic]
	if ok {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					printer.Error(err)
				}
			}()
			for _, s := range list {
				temp, ok := s.(*subscriber)
				if !ok {
					continue
				}
				temp.c <- event
			}
		}()
	}
}
