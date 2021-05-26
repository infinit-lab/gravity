package event

import (
	"errors"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	"sync"
	"time"
)

type Event struct {
	Topic   string      `json:"topic"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Context interface{} `json:"context,omitempty"`
}

type Subscriber interface {
	Unsubscribe()
}

func Subscribe(topic string, handler func(e *Event)) (Subscriber, error) {
	if handler == nil {
		return nil, errors.New("回调为空")
	}
	s := new(subscriber)
	s.c = make(chan *Event)
	s.topic = topic

	subscriberMutex.Lock()
	defer subscriberMutex.Unlock()

	list := subscriberMap[topic]
	list = append(list, s)
	subscriberMap[topic] = list

	go func() {
		for {
			e, ok := <-s.c
			if !ok {
				break
			}
			handler(e)
		}
	}()

	return s, nil
}

func SubscribeAll(handler func(e *Event)) (Subscriber, error) {
	return Subscribe("", handler)
}

func Publish(event *Event) error {
	if event == nil {
		return errors.New("The event is nil. ")
	}
	if len(event.Topic) == 0 {
		return errors.New("Topic is empty. ")
	}
	publishMutex.Lock()
	if idleWorkNum < 2 {
		publishEventCache = append(publishEventCache, event)
	} else {
		publishChan <- event
	}
	publishMutex.Unlock()
	return nil
}

var subscriberMap map[string][]Subscriber
var subscriberMutex sync.Mutex
var publishChan chan *Event
var idleWorkNum int
var publishEventCache []*Event
var publishMutex sync.Mutex

func init() {
	subscriberMap = make(map[string][]Subscriber)
	workerNum := config.GetInt("event.worker")
	if workerNum == 0 {
		workerNum = 30
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
				idleWorkNum++
				e := <-publishChan
				idleWorkNum--
				if idleWorkNum == 0 {
					printer.Trace("all event worker is used.")
				}
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

	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			publishMutex.Lock()
			if len(publishEventCache) != 0 {
				if idleWorkNum > 2 {
					publishChan <- publishEventCache[0]
					if len(publishEventCache) == 1 {
						publishEventCache = nil
					} else {
						publishEventCache = publishEventCache[1:]
					}
				}
			}
			publishMutex.Unlock()
		}
	}()
}

type subscriber struct {
	c     chan *Event
	topic string
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
