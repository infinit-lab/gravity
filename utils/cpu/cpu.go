package cpu

/*
#include <cpuid.h>
#include <stdio.h>
#include <string.h>

typedef struct _T_CPU_ID {
	char id[32];
}T_CPU_ID, *PT_CPU_ID;

void getCpuId(PT_CPU_ID id) {
	unsigned int cpuId[4] = {0};
	memset(id, 0, sizeof(T_CPU_ID));
	__cpuid(1, cpuId[0], cpuId[1], cpuId[2], cpuId[3]);
	sprintf(id->id, "%08X%08X", cpuId[3], cpuId[0]);
}
*/
import "C"
import (
	"github.com/infinit-lab/gravity/utils"
	"unsafe"
)

func GetCpuID() (string, error) {
	var id C.struct__T_CPU_ID
	C.getCpuId(&id)
	return utils.CArrayToGoString(unsafe.Pointer(&id.id[0]), 32), nil
}
