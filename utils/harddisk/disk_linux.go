// +build linux

package harddisk

import (
	"github.com/infinit-lab/gravity/printer"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
)

func GetDiskSerialNumber() ([]string, error) {
	cmd := exec.Command("blkid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		printer.Error("Failed to CombinedOutpu. error: ", err)
		return nil, err
	}
	blkMap := make(map[string]string)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		sections := strings.Split(line, ":")
		if len(sections) < 2 {
			continue
		}
		reg := regexp.MustCompile("UUID=\".*\" T")
		data := reg.Find([]byte(sections[1]))
		if len(data) == 0 {
			continue
		}
		uuid := strings.ReplaceAll(string(data), "UUID=\"", "")
		uuid = strings.ReplaceAll(uuid, "\" T", "")
		blkMap[sections[0]] = uuid
	}

	data, err := ioutil.ReadFile("/etc/fstab")
	if err != nil {
		printer.Error("Failed to ReadFile. error: ", err)
		return nil, err
	}
	lines = strings.Split(string(data), "\n")

	var uuidList []string
	for _, line := range lines {
		if strings.Contains(line, " / ") == false || strings.Contains(line, "#") {
			continue
		}
		sections := strings.Split(line, " / ")
		if len(sections) < 2 {
			continue
		}
		section := strings.ReplaceAll(sections[0], " ", "")
		if strings.Contains(section, "UUID=") {
			section = strings.ReplaceAll(section, "UUID=", "")
			uuidList = append(uuidList, section)
		} else {
			uuid, ok := blkMap[section]
			if ok {
				uuidList = append(uuidList, uuid)
			}
		}
	}
	return uuidList, nil
}
