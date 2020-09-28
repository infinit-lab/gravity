package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/server"
	"net/http"
)

type Response struct {
	Result bool        `json:"result"`
	Error  string      `json:"error"`
	Data   interface{} `json:"data,omitempty"`
}

type HandlerFunc func(context *gin.Context, session *Session) (interface{}, error)

type Controller interface {
	Group() *gin.RouterGroup

	GET(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	POST(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	PUT(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	DELETE(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	PATCH(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	HEAD(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	OPTIONS(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)

	GETWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	POSTWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	PUTWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	DELETEWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	PATCHWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	HEADWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
	OPTIONSWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc)
}

func New(prefix string) (Controller, error) {
	controller := new(controller)
	controller.group = server.Router().Group(prefix)
	return controller, nil
}

type controller struct {
	group *gin.RouterGroup
}

func (c *controller) Group() *gin.RouterGroup {
	return c.group
}

func (c *controller) GET(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.group.GET(relativePath, middleFunc(handler, middle)...)
}

func (c *controller) POST(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.group.POST(relativePath, middleFunc(handler, middle)...)
}

func (c *controller) PUT(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.group.PUT(relativePath, middleFunc(handler, middle)...)
}

func (c *controller) DELETE(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.group.DELETE(relativePath, middleFunc(handler, middle)...)
}

func (c *controller) PATCH(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.group.PATCH(relativePath, middleFunc(handler, middle)...)
}

func (c *controller) HEAD(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.group.HEAD(relativePath, middleFunc(handler, middle)...)
}

func (c *controller) OPTIONS(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.group.OPTIONS(relativePath, middleFunc(handler, middle)...)
}

func (c *controller) GETWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.GET(relativePath, handler, sessionMiddleFunc(middle)...)
}

func (c *controller) POSTWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.POST(relativePath, handler, sessionMiddleFunc(middle)...)
}

func (c *controller) PUTWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.PUT(relativePath, handler, sessionMiddleFunc(middle)...)
}

func (c *controller) DELETEWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.DELETE(relativePath, handler, sessionMiddleFunc(middle)...)
}

func (c *controller) PATCHWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.PATCH(relativePath, handler, sessionMiddleFunc(middle)...)
}

func (c *controller) HEADWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.HEAD(relativePath, handler, sessionMiddleFunc(middle)...)
}

func (c *controller) OPTIONSWithSession(relativePath string, handler HandlerFunc, middle ...gin.HandlerFunc) {
	c.OPTIONS(relativePath, handler, sessionMiddleFunc(middle)...)
}

func response(context *gin.Context, data interface{}, err error) {
	var status int
	var response Response
	if err == nil {
		response.Result = true
		response.Error = ""
		status = http.StatusOK
	} else {
		response.Result = false
		response.Error = err.Error()
		status = http.StatusInternalServerError
	}
	response.Data = data
	context.JSON(status, response)
}

func middleFunc(handler HandlerFunc, middle []gin.HandlerFunc) []gin.HandlerFunc {
	return append(middle, func(c *gin.Context) {
		token, ok := c.Get("Token")
		var session *Session
		if !ok {
			printer.Error("Not find token. ")
		} else {
			session, _ = GetSession(token.(string))
		}
		data, err := handler(c, session)
		response(c, data, err)
	})
}

func sessionMiddleFunc(middle []gin.HandlerFunc) []gin.HandlerFunc {
	var temp []gin.HandlerFunc
	temp = append(temp, SessionMiddle())
	temp = append(temp, middle...)
	return temp
}

func GetRequestBody(c *gin.Context, body interface{}) error {
	data, err := c.GetRawData()
	if err != nil {
		printer.Error(err)
		return err
	}
	return json.Unmarshal(data, body)
}
