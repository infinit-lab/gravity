package utils

import (
	"github.com/infinit-lab/gravity/printer"
	"golang.org/x/text/encoding/simplifiedchinese"
	"unsafe"
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

func ConvertGBKToUtf8(bytes []byte) string {
	decodeBytes, err := simplifiedchinese.GB18030.NewDecoder().Bytes(bytes)
	if err != nil {
		printer.Error("Failed to Bytes. error: ", err)
		return ""
	}
	return string(decodeBytes)
}
