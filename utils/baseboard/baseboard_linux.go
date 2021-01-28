// +build linux

package baseboard

import (
	"github.com/infinit-lab/yolanda/logutils"
	"os/exec"
	"strings"
)

func GetUUID() (string, error) {
	cmd := exec.Command("dmidecode", "--string", "system-uuid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		logutils.Error("Failed to CombineOutput. error: ", err)
		return "", err
	}
	line := strings.ReplaceAll(string(out), "\n", "")
	line = strings.ReplaceAll(line, "\r", "")
	line = strings.ReplaceAll(line, " ", "")
	return line, nil
}
