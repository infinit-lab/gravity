// +build windows

package baseboard

import (
	"github.com/infinit-lab/gravity/printer"
	"os/exec"
	"strings"
)

func GetUUID() (string, error) {
	cmd := exec.Command("wmic", "csproduct", "get", "uuid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		printer.Error("Failed to CombineOutput. error: ", err)
		return "", nil
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line != "" && strings.Contains(line, "UUID") == false {
			line = strings.ReplaceAll(line, " ", "")
			line = strings.ReplaceAll(line, "\r", "")
			return line, nil
		}
	}
	return "", nil
}
