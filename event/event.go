package event

import (
	"errors"
	"github.com/infinit-lab/gravity/config"
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
		return errors.New("The event is nil. ")
	}
	if len(event.Topic) == 0 {
		return errors.New("Topic is empty. ")
	}
	publishChan <- event
	return nil
}

var subscriberMap map[string][]Subscriber
var subscriberMutex sync.Mutex
var publishChan chan *Event

func init() {
	subscriberMap = make(map[string][]Subscriber)
	workerNum := config.GetInt("event.worker")
	if workerNum == 0 {
		workerNum = 20
	}
	publishChan = make(chan *Event, workerNum)
	for i := 0; i < workerNum; i++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					printer.Error(err)
				}
			}()
			for {
				e := <-publishChan
				var list []Subscriber
				subscriberMutex.Lock()
				topicList, ok := subscriberMap[e.Topic]
				if ok {
					list = append(list, topicList...)
				}
				allList, ok := subscriberMap[""]
				if ok {
					list = append(list, allList...)
				}
				subscriberMutex.Unlock()
				if ok {
					for _, s := range list {
						temp := s.(*subscriber)
						temp.c <- e
					}
				}
			}
		}()
	}
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
