package config

import (
	"github.com/infinit-lab/gravity/printer"
)

var reader yamlReader

func init() {
	err := reader.load()
	if err != nil {
		printer.Error(err)
	}
	reader.loadArgs()
}

func LoadArgs() {
	reader.loadArgs()
}

func GetString(name string) string {
	return reader.getString(name)
}

func GetInt(name string) int {
	return reader.getInt(name)
}

func GetBool(name string) bool {
	return reader.getBool(name)
}

func GetFloat64(name string) float64 {
	return reader.getFloat64(name)
}
