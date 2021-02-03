package snow_flack

import (
	"github.com/infinit-lab/gravity/printer"
	"os"
)

var worker *IdWorker

func init() {
	var err error
	worker, err = NewIdWorker(1)
	if err != nil {
		printer.Error(err)
		os.Exit(1)
	}
}

func NextId() (int64, error) {
	return worker.NextId()
}

func NextIds(num int) ([]int64, error) {
	return worker.NextIds(num)
}
