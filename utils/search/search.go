package search

import (
	"encoding/json"
	"fmt"
	"github.com/infinit-lab/gravity/activation"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils/network"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Request struct {
	Command string `json:"command"`
	Session int    `json:"session"`
}

type AuthRequest struct {
	Request
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Request
	Result bool   `json:"result"`
	Error  string `json:"error"`
}

type udpFrameHandler struct {
	key string
}

type udpClient struct {
	cache      *cache
	timer      *time.Timer
	handler    *udpFrameHandler
	addr       *net.UDPAddr
	frameIndex uint16
}

type udpServer struct {
	conn       *net.UDPConn
	cacheMap   map[string]*udpClient
	cacheMutex sync.Mutex
}

const (
	cmdSearch       string = "search"
	cmdNetList      string = "net_list"
	cmdSetNet       string = "set_net"
	cmdUpdate       string = "update"
	cmdUpdateNotify string = "update_notify"
)

var server *udpServer

type Version struct {
	Version string `json:"version"`
	CommitId string `json:"commitId"`
	BuildTime string `json:"buildTime"`
}
var version Version

type License struct {
	Fingerprint string `json:"fingerprint"`
	Status string `json:"status"`
	IsForever bool `json:"isForever"`
	ValidDatetime string `json:"validDatetime"`
	ValidDuration int `json:"validDuration"`
}

func init() {
	server = new(udpServer)
	server.cacheMap = make(map[string]*udpClient)

	port := config.GetInt("search.port")
	printer.Trace("Get search.port is ", port)
	if port == 0 {
		port = 5254
		printer.Trace("Reset search.port to ", port)
	}
	address := "0.0.0.0:" + strconv.Itoa(port)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		printer.Error("Failed to ResolveUDPAddr. error: ", err)
		os.Exit(1)
	}

	server.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		printer.Error("Failed to ListenUDP. error: ", err)
		os.Exit(1)
	}

	bufferSize := config.GetInt("search.bufferSize")
	printer.Trace("Get search.bufferSize is ", bufferSize)
	if bufferSize == 0 {
		bufferSize = 1024
		printer.Trace("Reset search.bufferSize to ", bufferSize)
	}

	go func() {
		defer func() {
			_ = server.conn.Close()
		}()

		for {
			data := make([]byte, bufferSize)
			recvLen, rAddr, err := server.conn.ReadFromUDP(data)
			if err != nil {
				printer.Error("Failed to ReadFromUDP")
				break
			}
			key := fmt.Sprintf("%d.%d.%d.%d:%d", rAddr.IP[0], rAddr.IP[1],
				rAddr.IP[2], rAddr.IP[3], rAddr.Port)
			server.cacheMutex.Lock()
			cache, ok := server.cacheMap[key]
			if !ok {
				cache = new(udpClient)
				cache.handler = new(udpFrameHandler)
				cache.handler.key = key
				cache.cache = newCache(cache.handler)
				cache.timer = time.NewTimer(time.Second * 30)
				cache.addr = rAddr
				go func() {
					select {
					case <-cache.timer.C:
						cache.cache.close()
						server.cacheMutex.Lock()
						delete(server.cacheMap, key)
						server.cacheMutex.Unlock()
					}
				}()
				server.cacheMap[key] = cache
			}
			server.cacheMutex.Unlock()
			if cache != nil {
				cache.timer.Reset(time.Second * 30)
				cache.cache.push(data[:recvLen])
			}
		}
	}()
}

type searchResponse struct {
	Response
	Data struct {
		Version Version  `json:"version"`
		License License `json:"license"`
	} `json:"data"`
}

type netListResponse struct {
	Response
	Data []*network.Adapter `json:"data"`
}

type setNetRequest struct {
	Request
	Data network.Adapter `json:"data"`
}

func SetVersion(v Version) {
	version = v
}

var authorizationFunc func(string, string) error

func SetAuthorizationFunction(f func(string, string) error) {
	authorizationFunc = f
}

func checkAccount(buffer []byte) (bool, error) {
	var user AuthRequest
	err := json.Unmarshal(buffer, &user)
	if err != nil {
		printer.Error("Failed to Unmarshal. error: ", err)
		return false, err
	}
	if authorizationFunc == nil {
		return true, nil
	}
	err = authorizationFunc(user.Username, user.Password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (h *udpFrameHandler) onGetFrame(buffer []byte) {
	var request Request
	err := json.Unmarshal(buffer, &request)
	if err != nil {
		printer.Error("Failed to Unmarshal. error: ", err)
		return
	}
	switch request.Command {
	case cmdSearch:
		var response searchResponse
		response.Request = request
		var err error
		response.Data.License.IsForever, response.Data.License.ValidDatetime, response.Data.License.ValidDuration, err = activation.GetLicenseDuration()
		if err != nil {
			printer.Error("Failed to GetMachineFingerprint. error: ", err)
			h.responseError(request, err.Error())
			return
		}
		response.Data.License.Fingerprint = activation.GetFingerprint()
		response.Data.License.Status = activation.GetStatus()
		response.Data.Version = version
		response.Result = true
		h.response(response)
	case cmdNetList:
		isValid, err := checkAccount(buffer)
		if err != nil {
			h.responseError(request, err.Error())
			return
		}
		if !isValid {
			h.responseError(request, "无效用户名或密码")
			return
		}
		var response netListResponse
		response.Request = request
		response.Data, err = network.GetNetworkInfo()
		if err != nil {
			printer.Error("Failed to GetNetworkInfo. error: ", err)
			h.responseError(request, err.Error())
			return
		}
		response.Result = true
		h.response(response)
	case cmdSetNet:
		isValid, err := checkAccount(buffer)
		if err != nil {
			h.responseError(request, err.Error())
			return
		}
		if !isValid {
			h.responseError(request, "无效用户名或密码")
			return
		}
		var req setNetRequest
		err = json.Unmarshal(buffer, &req)
		if err != nil {
			printer.Error("Failed to Unmarshal. error: ", err)
			h.responseError(request, err.Error())
			return
		}
		err = network.SetAdapter(&req.Data)
		if err != nil {
			printer.Error("Failed to SetAdapter. error: ", err)
			h.responseError(request, err.Error())
			return
		}
		var response Response
		response.Request = request
		response.Result = true
		h.response(response)
	case cmdUpdate:
	default:
		return
	}
}

func (s *udpServer) getClientByHandler(h *udpFrameHandler) *udpClient {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	client, ok := s.cacheMap[h.key]
	if !ok {
		return nil
	}
	return client
}

func (h *udpFrameHandler) responseError(request Request, err string) {
	var response Response
	response.Request = request
	response.Result = false
	response.Error = err
	h.response(response)
}

func (h *udpFrameHandler) response(response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		printer.Error("Failed to Marshal. error: ", err)
		return
	}
	printer.Trace(string(data))
	client := server.getClientByHandler(h)
	if client == nil {
		printer.Error("Failed to getAddrByHandler")
		return
	}
	printer.Trace(client.addr.IP.String())
	printer.Trace(client.addr.Port)
	client.frameIndex++
	frameList := packBuffer(data, client.frameIndex)
	for _, frame := range frameList {
		_, _ = server.conn.WriteToUDP(frame, client.addr)
	}
}
