package notifier

import (
	"github.com/infinit-lab/gravity/model"
)

type TestResource struct {
	model.Resource
	Description string `json:"description" db:"description"`
}

/*
var id int

func TestNotifier(t *testing.T) {
	_, err := module.New("/api", "test", &TestResource{})
	if err != nil {
		t.Fatal(err)
	}
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

	subscriber, err := event.Subscribe("test")
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			e, ok := <-subscriber.Event()
			if !ok {
				break
			}
			value, _ := e.Data.(model.Id)
			id = value.GetId()
			data, _ := json.Marshal(e)
			printer.Trace(string(data))
		}
	}()

	session, err := controller.CreateSession(1, "User", "127.0.0.1", nil)
	if err != nil {
		t.Fatal(err)
	}

	u := url.URL{
		Scheme:   "ws",
		Host:     "127.0.0.1:8081",
		Path:     "/ws/notification",
		RawQuery: "token=" + session.Token,
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			printer.Trace(string(message))
		}
	}()

	time.Sleep(10 * time.Millisecond)

	var tr TestResource
	tr.Name = "Test"
	tr.Description = "TestDescription"
	data, _ := json.Marshal(tr)
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8081/api/data", strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	time.Sleep(10 * time.Millisecond)
	if id == 0 {
		t.Fatal("id is zero")
	}

	if req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/api/data", nil); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	if req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/api/data/"+fmt.Sprint(id), nil); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	tr.Id = id
	tr.Name = "UpdateTest"
	tr.Description = "UpdateDescription"
	data, _ = json.Marshal(tr)

	time.Sleep(2 * time.Second)

	if req, err = http.NewRequest(http.MethodPut, "http://127.0.0.1:8081/api/data/"+fmt.Sprint(id), strings.NewReader(string(data))); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	if req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/api/data", nil); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	if req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/api/data/"+fmt.Sprint(id), nil); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	if req, err = http.NewRequest(http.MethodDelete, "http://127.0.0.1:8081/api/data/"+fmt.Sprint(id), nil); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	if req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/api/data", nil); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	if req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/api/data/"+fmt.Sprint(id), nil); err != nil {
		t.Fatal(err)
	}
	req.Header["Authorization"] = []string{session.Token}
	if rsp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatal(err)
	}
	if body, err = ioutil.ReadAll(rsp.Body); err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()
	printer.Trace(string(body))

	time.Sleep(10 * time.Millisecond)
	_ = conn.Close()
	subscriber.Unsubscribe()
	_ = server.Shutdown()
	wg.Wait()
}

*/
