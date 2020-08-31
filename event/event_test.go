package event

import (
	"encoding/json"
	"log"
	"sync"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	var wg sync.WaitGroup

	all, err := SubscribeAll()
	if err != nil {
		t.Fatal("Failed to SubscribeAll. error: ", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			event, ok := <-all.Event()
			if !ok {
				break
			}
			data, _ := json.Marshal(event)
			log.Print("All:", string(data))
			if event.Status == "stop" {
				all.Unsubscribe()
			}
		}
	}()

	single, err := Subscribe("test")
	if err != nil {
		t.Fatal("Failed to Subscribe. error: ", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			event, ok := <-single.Event()
			if !ok {
				break
			}
			data, _ := json.Marshal(event)
			log.Print("Single:", string(data))
			if event.Status == "stop" {
				single.Unsubscribe()
			}
		}
	}()

	event1 := Event{
		Topic:   "test",
		Status:  "event1",
		Data:    "event1",
		Context: "event1",
	}
	_ = Publish(&event1)

	time.Sleep(1 * time.Millisecond)

	event2 := Event{
		Topic:   "event2",
		Status:  "event2",
		Data:    "event2",
		Context: "event2",
	}
	_ = Publish(&event2)

	time.Sleep(1 * time.Millisecond)

	event3 := Event{
		Topic:   "event3",
		Status:  "stop",
		Data:    "event3",
		Context: "event3",
	}
	_ = Publish(&event3)

	time.Sleep(1 * time.Millisecond)

	event4 := Event{
		Topic:   "test",
		Status:  "stop",
		Data:    "event4",
		Context: "event4",
	}
	_ = Publish(&event4)

	wg.Wait()
}
