// +build windows

package harddisk

import (
	"github.com/infinit-lab/gravity/printer"
	"os/exec"
	"strings"
)

func GetDiskSerialNumber() ([]string, error) {
	cmd := exec.Command("wmic", "diskdrive", "get", "serialnumber")
	out, err := cmd.CombinedOutput()
	if err != nil {
		printer.Error("Failed to CombinedOutput. error: ", err)
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var serialNumbers []string
	for _, line := range lines {
		line = strings.ReplaceAll(line, " ", "")
		line = strings.ReplaceAll(line, "\r", "")
		if line != "" && strings.Contains(line, "SerialNumber") == false {
			serialNumbers = append(serialNumbers, line)
		}
	}
	if len(serialNumbers) == 0 {
		serialNumbers = append(serialNumbers, "")
	}
	return serialNumbers, nil
}
