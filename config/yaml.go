package config

import (
	"fmt"
	"github.com/infinit-lab/gravity/printer"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type yamlReader struct {
	data map[interface{}]interface{}
}

func isPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (r *yamlReader) loadArgs() {
	if r.data == nil {
		r.data = make(map[interface{}]interface{})
	}
	for _, v := range os.Args {
		if !strings.Contains(v, "=") {
			continue
		}
		index := strings.Index(v, "=")
		r.set(v[:index], v[index+1:])
	}
	r.print()
}

func (r *yamlReader) print() {
	r.printData(r.data, 0)
}

func (r *yamlReader) printData(data map[interface{}]interface{}, depth int) {
	var tab string
	for i := 0; i < depth; i++ {
		tab += " "
	}
	for key, value := range data {
		if value == nil {
			continue
		}
		if reflect.TypeOf(value).String() == "map[interface {}]interface {}" {
			printer.Tracef(tab+"%v:", key)
			r.printData(value.(map[interface{}]interface{}), depth+1)
		} else {
			printer.Tracef(tab+"%v:%v", key, value)
		}
	}
}

func (r *yamlReader) load() error {
	const path = "config.yaml"
	isExists, err := isPathExists(path)
	if err != nil {
		return err
	}
	if !isExists {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		_ = f.Close()
	}
	stream, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(stream, &r.data)
	return err
}

func (r *yamlReader) get(name string) interface{} {
	path := strings.Split(name, ".")
	data := r.data
	for key, value := range path {
		v, ok := data[value]
		if !ok {
			break
		}
		if (key + 1) == len(path) {
			return v
		}
		if reflect.TypeOf(v).String() == "map[interface {}]interface {}" {
			data = v.(map[interface{}]interface{})
		}
	}
	return nil
}

func (r *yamlReader) set(name string, input interface{}) {
	path := strings.Split(name, ".")
	data := r.data
	for key, value := range path {
		if (key + 1) == len(path) {
			data[value] = input
			return
		}
		v, ok := data[value]
		if !ok {
			v = make(map[interface{}]interface{})
			data[value] = v
		}
		if reflect.TypeOf(v).String() == "map[interface {}]interface {}" {
			data = v.(map[interface{}]interface{})
		} else {
			return
		}
	}
}

func (r *yamlReader) getString(name string) string {
	value := r.get(name)
	switch value := value.(type) {
	case string:
		return value
	case bool, float64, int:
		return fmt.Sprint(value)
	default:
		return ""
	}
}

func (r *yamlReader) getInt(name string) int {
	value := r.get(name)
	switch value := value.(type) {
	case string:
		i, _ := strconv.Atoi(value)
		return i
	case int:
		return value
	case bool:
		if value {
			return 1
		}
		return 0
	case float64:
		return int(value)
	default:
		return 0
	}
}

func (r *yamlReader) getBool(name string) bool {
	value := r.get(name)
	switch value := value.(type) {
	case string:
		str, _ := strconv.ParseBool(value)
		return str
	case int:
		if value != 0 {
			return true
		}
		return false
	case bool:
		return value
	case float64:
		if value != 0.0 {
			return true
		}
		return false
	default:
		return false
	}
}

func (r *yamlReader) getFloat64(name string) float64 {
	value := r.get(name)
	switch value := value.(type) {
	case string:
		str, _ := strconv.ParseFloat(value, 64)
		return str
	case int:
		return float64(value)
	case bool:
		if value {
			return float64(1)
		}
		return float64(0)
	case float64:
		return value
	default:
		return 0.0
	}
}
