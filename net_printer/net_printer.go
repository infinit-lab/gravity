package net_printer

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/server"
	"net/http"
	"strconv"
	"sync"
)

type printerHandler struct {
	socket      server.Websocket
	socketMutex sync.Mutex
}

func (h *printerHandler) NewConnection(socket server.Websocket) {
	h.socketMutex.Lock()
	defer h.socketMutex.Unlock()
	if h.socket != nil {
		_ = h.socket.Close()
	}
	h.socket = socket
}

func (h *printerHandler) Disconnected(socket server.Websocket) {
	h.socketMutex.Lock()
	defer h.socketMutex.Unlock()
	if h.socket == socket {
		h.socket = nil
	}
}

func (h *printerHandler) ReadMessage(socket server.Websocket, message []byte) {

}

func (h *printerHandler) ReadBytes(socket server.Websocket, bytes []byte) {

}

func (h *printerHandler) Write(p []byte) (int, error) {
	h.socketMutex.Lock()
	defer h.socketMutex.Unlock()
	if h.socket != nil {
		_ = h.socket.Socket().WriteMessage(websocket.TextMessage, p)
	}
	return len(p), nil
}

var handler *printerHandler

func init() {
	handler = new(printerHandler)
	printer.RegisterWriter(handler)
	server.Router().GET("/ws/printer", server.GenerateWebsocketHandlerFunc(handler))
	server.Router().GET("/api/printer/level/:level", func(c *gin.Context) {
		level, err := strconv.Atoi(c.Param("level"))
		if err != nil {
			printer.Error(err)
			c.JSON(http.StatusBadRequest, nil)
			return
		}
		printer.SetLevel(level)
		c.JSON(http.StatusOK, nil)
		return
	})
	server.Router().GET("/view/printer", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Type", "text/html")
		_, _ = c.Writer.Write([]byte(handler.GetPrinterView()))
	})
}

func (h *printerHandler) GetPrinterView() string {
	return "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\"><title>Printer</title><script>" +
		"let url = \"ws://\" + window.location.host + \"/ws/printer\";\n" +
		"let ws = new WebSocket(url);\n" +
		"ws.onopen = function() {\n" +
		"\tconsole.log(\"Connected...\");\n" +
		"};\n" +
		"ws.onclose = function() {\n" +
		"\tconsole.log(\"Disconnected. Please refresh.\");\n" +
		"};\n" +
		"ws.onmessage = function (e) {\n" +
		"\tif (e.data.indexOf(\"ERROR:\") !== -1) {\n" +
		"\t\tconsole.error(e.data)\n" +
		"\t} else if (e.data.indexOf(\"WARNING:\") !== -1) {\n" +
		"\t\tconsole.info(e.data)\n" +
		"\t} else {\n" +
		"\t\tconsole.log(e.data)\n" +
		"\t}\n" +
		"}\n" +
		"</script></head></html>"
}
