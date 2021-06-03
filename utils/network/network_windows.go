package network

/*
#cgo CFLAGS: -DWINDOWS -O3
#cgo LDFLAGS: -liphlpapi
#include "network.h"
*/
import "C"
import (
	"github.com/infinit-lab/gravity/printer"
	"net"
	"os/exec"
	"strings"
	"unsafe"
)

func GetNetworkInfo() ([]*Adapter, error) {
	var adapters []*Adapter
	count := C.GetAdaptersCount()
	for i := C.int(0); i < count; i++ {
		var adapter C.struct__T_ADAPTER
		ret := C.GetAdapter(i, &adapter)
		if ret == 0 {
			a := new(Adapter)
			a.Index = int(adapter.index)
			a.Name = CArrayToGoString(unsafe.Pointer(&adapter.name[0]), nameLength)
			a.Description = CArrayToGoString(unsafe.Pointer(&adapter.description[0]), nameLength)
			a.Mac = CArrayToGoString(unsafe.Pointer(&adapter.mac[0]), addrLength)
			a.Ip = CArrayToGoString(unsafe.Pointer(&adapter.ip[0]), addrLength)
			a.Mask = CArrayToGoString(unsafe.Pointer(&adapter.mask[0]), addrLength)
			a.Gateway = CArrayToGoString(unsafe.Pointer(&adapter.gateway[0]), addrLength)
			if strings.Contains(a.Description, "Virtual") {
				continue
			}
			adapters = append(adapters, a)
		}
	}
	for _, adapter := range adapters {
		i, err := net.InterfaceByIndex(adapter.Index)
		if err != nil {
			printer.Error("Failed to InterfaceByIndex. error: ", err)
			return nil, err
		}
		adapter.Name = i.Name
	}
	return adapters, nil
}

func SetAdapter(adapter *Adapter) error {
	cmd := new(exec.Cmd)
	cmd.Path = "netsh"
	cmd.Args = []string{
		"netsh",
		"interface",
		"ip",
		"set",
		"address",
		"\"" + adapter.Name + "\"",
		"static",
		adapter.Ip,
		adapter.Mask,
		adapter.Gateway,
		"1",
	}
	if err := cmd.Run(); err != nil {
		printer.Error("Failed to Run. error: ", err)
		return err
	}
	return nil
}
