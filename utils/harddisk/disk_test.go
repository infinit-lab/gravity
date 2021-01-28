package harddisk

import (
	"log"
	"testing"
)

func TestGetDiskSerialNumber(t *testing.T) {
	serialNumbers, err := GetDiskSerialNumber()
	if err != nil {
		t.Fatal("Failed to GetDiskSerialNumber. error: ", err)
	}
	for _, serialNumber := range serialNumbers {
		log.Println(serialNumber)
	}
}
