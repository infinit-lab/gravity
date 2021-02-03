package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	"net/http"
	"time"
)

func Router() *gin.Engine {
	return router
}

func Run() error {
	defer func() {
		server = nil
		fileHandler = nil
	}()
	port := config.GetInt("server.port")
	if port == 0 {
		port = 8080
	}
	server = new(http.Server)
	server.Addr = fmt.Sprintf(":%d", port)
	server.Handler = router

	assets := config.GetString("server.assets")
	if len(assets) == 0 {
		assets = "./assets"
	}
	fileHandler = http.FileServer(http.Dir(assets))
	router.NoRoute(func(context *gin.Context) {
		if fileHandler != nil {
			fileHandler.ServeHTTP(context.Writer, context.Request)
		} else {
			context.JSON(http.StatusNotFound, gin.H{"result": false, "message": "Not Found. "})
		}
	})
	return server.ListenAndServe()
}

func Shutdown() error {
	if server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	return server.Shutdown(ctx)
}

func GetAssetsPath() string {
	assets := config.GetString("server.assets")
	if len(assets) == 0 {
		assets = "./assets"
	}
	return assets
}

type Websocket interface {
	Socket() *websocket.Conn
	Context() *gin.Context
	WriteMessage(message []byte) error
	WriteBytes(bytes []byte) error
	Close() error
}

type WebsocketHandler interface {
	NewConnection(socket Websocket)
	Disconnected(socket Websocket)
	ReadMessage(socket Websocket, message []byte)
	ReadBytes(socket Websocket, bytes []byte)
}

func GenerateWebsocketHandlerFunc(handler WebsocketHandler) gin.HandlerFunc {
	return func(context *gin.Context) {
		if !context.IsWebsocket() {
			printer.Error("Scheme isn't websocket")
			return
		}
		ws, err := upgrader.Upgrade(context.Writer, context.Request, nil)
		if err != nil {
			printer.Error(err)
			return
		}
		w := websocketImpl{
			ws:               ws,
			context:          context,
			isClose:          false,
			writeMessageChan: make(chan []byte),
			writeBytesChan:   make(chan []byte),
		}

		go func() {
			defer func() {
				if err := recover(); err != nil {
					printer.Error(err)
				}
			}()
			for !w.isClose {
				select {
				case message, ok := <-w.writeMessageChan:
					if !ok {
						return
					}
					if err := w.ws.WriteMessage(websocket.TextMessage, message); err != nil {
						printer.Error(err)
					}
				case bytes, ok := <-w.writeBytesChan:
					if !ok {
						return
					}
					if err := w.ws.WriteMessage(websocket.BinaryMessage, bytes); err != nil {
						printer.Error(err)
					}
				}
			}
		}()
		handler.NewConnection(&w)
		for {
			messageType, bytes, err := w.ws.ReadMessage()
			if err != nil {
				break
			}
			switch messageType {
			case websocket.TextMessage:
				go handler.ReadMessage(&w, bytes)
			case websocket.BinaryMessage:
				go handler.ReadBytes(&w, bytes)
			}
		}
		w.isClose = true
		close(w.writeMessageChan)
		close(w.writeBytesChan)
		_ = w.ws.Close()
		handler.Disconnected(&w)
	}
}

var router *gin.Engine
var server *http.Server
var upgrader websocket.Upgrader
var fileHandler http.Handler

func init() {
	gin.SetMode(gin.ReleaseMode)
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	router = gin.Default()
	router.Use(func(context *gin.Context) {
		isPrintAccess := config.GetBool("server.print_access")
		if isPrintAccess {
			printer.Trace(context.Request.RemoteAddr, context.Request.Method, context.Request.URL.String())
		}
	})
}

type websocketImpl struct {
	ws               *websocket.Conn
	context          *gin.Context
	isClose          bool
	writeMessageChan chan []byte
	writeBytesChan   chan []byte
}

func (w *websocketImpl) Socket() *websocket.Conn {
	return w.ws
}

func (w *websocketImpl) Context() *gin.Context {
	return w.context
}

func (w *websocketImpl) WriteMessage(message []byte) error {
	if w.isClose {
		return errors.New("The websocket is closed. ")
	}
	defer func() {
		if err := recover(); err != nil {
			printer.Error(err)
		}
	}()
	w.writeMessageChan <- message
	return nil
}

func (w *websocketImpl) WriteBytes(bytes []byte) error {
	if w.isClose {
		return errors.New("The websocket is closed. ")
	}
	defer func() {
		if err := recover(); err != nil {
			printer.Error(err)
		}
	}()
	w.writeBytesChan <- bytes
	return nil
}

func (w *websocketImpl) Close() error {
	if w.isClose {
		return errors.New("The websocket is closed. ")
	}
	return w.ws.Close()
}
