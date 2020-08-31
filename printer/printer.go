package printer

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
)

const (
	LevelError   int = 0x00000001
	LevelWarning int = 0x00000002 | LevelError
	LevelTrace   int = 0x00000004 | LevelWarning
)

func RegisterWriter(writer io.Writer) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.writers = append(w.writers, writer)
}

func SetLevel(l int) {
	level = l
}

func Trace(v ...interface{}) {
	print(LevelTrace, v...)
}

func Tracef(format string, args ...interface{}) {
	printf(LevelTrace, format, args...)
}

func Warning(v ...interface{}) {
	print(LevelWarning, v...)
}

func Warningf(format string, args ...interface{}) {
	printf(LevelWarning, format, args...)
}

func Error(v ...interface{}) {
	print(LevelError, v...)
}

func Errorf(format string, args ...interface{}) {
	printf(LevelError, format, args...)
}

type writer struct {
	writers []io.Writer
	mutex   sync.Mutex
}

func (w *writer) Write(p []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, writer := range w.writers {
		_, _ = writer.Write(p)
	}
	return len(p), nil
}

var (
	trace   *log.Logger
	warning *log.Logger
	err     *log.Logger
	level   int
	w       *writer
)

func init() {
	w = new(writer)
	trace = log.New(io.MultiWriter(os.Stdout, w), "TRACE: ", log.Ldate|log.Ltime)
	warning = log.New(io.MultiWriter(os.Stdout, w), "WARNING: ", log.Ldate|log.Ltime)
	err = log.New(io.MultiWriter(os.Stderr, w), "ERROR: ", log.Ldate|log.Ltime)
	level = LevelTrace
}

func caller() string {
	_, file, line, ok := runtime.Caller(3)
	if ok {
		_, fileName := filepath.Split(file)
		return fileName + ":" + strconv.Itoa(line)
	}
	return ""
}

func print(l int, v ...interface{}) {
	var value []interface{}
	value = append(value, caller())
	value = append(value, v...)
	if level&l == l {
		switch l {
		case LevelError:
			err.Println(value...)
		case LevelWarning:
			warning.Println(value...)
		case LevelTrace:
			trace.Println(value...)
		}
	}
}

func printf(l int, format string, args ...interface{}) {
	var value []interface{}
	value = append(value, caller())
	value = append(value, args...)
	format = "%s " + format
	if level&l == l {
		switch l {
		case LevelError:
			err.Printf(format, value...)
		case LevelWarning:
			warning.Printf(format, value...)
		case LevelTrace:
			trace.Printf(format, value...)
		}
	}
}
