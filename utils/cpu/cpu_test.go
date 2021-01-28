package cpu

import (
	"log"
	"testing"
)

func TestGetCpuUseRate(t *testing.T) {
	rate, err := GetCpuUseRate()
	if err != nil {
		t.Error("Failed to GetCpuUseRate. error: ", err)
	}
	log.Print("GetCpuUseRate: ", rate)
	cpuId, err := GetCpuID()
	if err != nil {
		t.Error("Failed to GetCpuID. error: ", err)
	}
	log.Print("GetCpuID: ", cpuId)
}
