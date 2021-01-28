// +build linux

package cpu

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func procStat() (idle, total uint64, err error) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(string(contents), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					return 0, 0, err
				}
				total += val
				if i == 4 {
					idle = val
				}
			}
			return
		}
	}
	return 0, 0, errors.New("文件格式错误")
}

func GetCpuUseRate() (int, error) {
	oldIdle, oldTotal, err := procStat()
	if err != nil {
		return 0, err
	}
	time.Sleep(time.Second)
	idle, total, err := procStat()
	if err != nil {
		return 0, err
	}
	idleTicks := idle - oldIdle
	totalTicks := total - oldTotal
	useRate := 100 * (totalTicks - idleTicks) / totalTicks
	return int(useRate), nil
}
