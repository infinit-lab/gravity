package network

import "unsafe"

type Adapter struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Mac         string `json:"mac"`
	Ip          string `json:"ip"`
	Mask        string `json:"mask"`
	Gateway     string `json:"gateway"`
}

const (
	nameLength = 256
	addrLength = 32
)

func CArrayToGoString(cArray unsafe.Pointer, size int) string {
	var goArray []byte
	p := uintptr(cArray)
	for i := 0; i < size; i++ {
		j := *(*byte)(unsafe.Pointer(p))
		if j == 0 {
			break
		}
		goArray = append(goArray, j)
		p += unsafe.Sizeof(j)
	}
	return string(goArray)
}

func GoStringToCArray(str string, cArray unsafe.Pointer, size int) {
	goArray := []byte(str)
	p := uintptr(cArray)
	for i := 0; i < size && i < len(goArray); i++ {
		*(*byte)(unsafe.Pointer(p)) = goArray[i]
		p += unsafe.Sizeof(goArray[i])
	}
}
