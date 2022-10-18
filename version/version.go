package version

import "fmt"

var MajorMinor string
var Date string
var Hash string

func GetVersion() string {
	return fmt.Sprintf("%s.%s.%s", MajorMinor, Hash, Date)
}
