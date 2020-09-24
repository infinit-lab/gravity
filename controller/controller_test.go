package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/server"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	c, err := New("/api")
	if err != nil {
		t.Fatal(err)
	}

	f := func(c *gin.Context, s *Session) (interface{}, error) {
		resource := database.Resource{
			PrimaryKey: database.PrimaryKey{
				Id: 1,
			},
			Name:    "Test",
			Creator: "TestCreator",
			Updater: "TestUpdater",
		}
		return resource, nil
	}

	c.GET("/test", f)
	c.POST("/test", f)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		os.Args = append(os.Args, "server.port=8081")
		config.LoadArgs()
		if err := server.Run(); err != nil {
			printer.Error(err)
		}
	}()

	resp, err := http.DefaultClient.Get("http://127.0.0.1:8081/api/test")
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	printer.Trace(string(body))

	resp, err = http.Post("http://127.0.0.1:8081/api/test", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	printer.Trace(string(body))

	if err := server.Shutdown(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func TestSession(t *testing.T) {
	subscriber, err := event.Subscribe(TopicSession)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			e, ok := <-subscriber.Event()
			if !ok {
				break
			}
			data, _ := json.Marshal(e)
			printer.Trace(string(data))
			if e.Status == model.StatusDeleted {
				subscriber.Unsubscribe()
			}
		}
	}()

	os.Args = append(os.Args, "session.age=2")
	config.LoadArgs()
	session, err := CreateSession("1", "tester", "127.0.0.1", nil)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	UpdateSession(session.Token)

	time.Sleep(time.Second)
	UpdateSession(session.Token)
	wg.Wait()
}

func TestControllerSessionMiddle(t *testing.T) {
	c, err := New("/session")
	if err != nil {
		t.Fatal(err)
	}

	f := func(c *gin.Context, s *Session) (interface{}, error) {
		resource := database.Resource{
			PrimaryKey: database.PrimaryKey{
				Id: 1,
			},
			Name:    s.Token,
			Creator: "TestCreator",
			Updater: "TestUpdater",
		}
		return resource, nil
	}

	c.GETWithSession("/test", f)
	c.POSTWithSession("/test", f)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		os.Args = append(os.Args, "server.port=8081")
		config.LoadArgs()
		if err := server.Run(); err != nil {
			printer.Error(err)
		}
	}()

	resp, err := http.DefaultClient.Get("http://127.0.0.1:8081/session/test")
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	printer.Trace(string(body))

	resp, err = http.Post("http://127.0.0.1:8081/session/test", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	printer.Trace(string(body))
	session, err := CreateSession("1", "tester", "127.0.0.1", nil)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/session/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	printer.Trace(string(body))

	req, err = http.NewRequest(http.MethodPost, "http://127.0.0.1:8081/session/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	printer.Trace(string(body))

	if err := server.Shutdown(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
