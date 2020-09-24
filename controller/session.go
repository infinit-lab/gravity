package controller

import (
	"errors"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/model"
	uuid "github.com/satori/go.uuid"
	"strings"
	"sync"
	"time"
)

const DefaultAge int = 600

type Session struct {
	Token     string      `json:"token"`
	UserId    string      `json:"userId"`
	Username  string      `json:"username"`
	Ip        string      `json:"ip"`
	Context   interface{} `json:"context,omitempty"`
	timer     *time.Timer
	closeChan chan int
}

var sessionMap map[string]*Session
var sessionMutex sync.Mutex

const (
	TopicSession string = "session"
)

func CreateSession(userId string, username string, ip string, context interface{}) (*Session, error) {
	session := new(Session)
	session.Token = strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	session.UserId = userId
	session.Username = username
	session.Ip = ip
	session.Context = context
	age := config.GetInt("session.age")
	if age == 0 {
		age = DefaultAge
	}
	session.timer = time.NewTimer(time.Duration(age) * time.Second)
	session.closeChan = make(chan int)
	go func() {
		select {
		case <-session.timer.C:
			DeleteSession(session.Token)
			break
		case <-session.closeChan:
			break
		}
	}()

	sessionMutex.Lock()
	sessionMap[session.Token] = session
	sessionMutex.Unlock()

	e := new(event.Event)
	e.Topic = TopicSession
	e.Status = model.StatusCreated
	e.Data = session
	e.Context = nil
	_ = event.Publish(e)
	return session, nil
}

func DeleteSession(token string) {
	sessionMutex.Lock()
	session, ok := sessionMap[token]
	if ok {
		delete(sessionMap, token)
	}
	sessionMutex.Unlock()
	if !ok {
		return
	}

	session.timer.Stop()
	close(session.closeChan)
	e := new(event.Event)
	e.Topic = TopicSession
	e.Status = model.StatusDeleted
	e.Data = session
	e.Context = nil
	_ = event.Publish(e)
}

func UpdateSession(token string) {
	sessionMutex.Lock()
	session, ok := sessionMap[token]
	sessionMutex.Unlock()
	if !ok {
		return
	}
	age := config.GetInt("session.age")
	if age == 0 {
		age = DefaultAge
	}
	session.timer.Reset(time.Duration(age) * time.Second)
}

func GetSession(token string) (*Session, error) {
	sessionMutex.Lock()
	session, ok := sessionMap[token]
	sessionMutex.Unlock()
	if !ok {
		return nil, errors.New("Invalid token. ")
	}
	return session, nil
}

func init() {
	sessionMap = make(map[string]*Session)
}
