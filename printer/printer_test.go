package printer

import "testing"

func TestPrinter(t *testing.T) {
	SetLevel(LevelTrace)
	Trace("This is a trace. ", "This is a trace")
	Warning("This is a warning. ", "This is a warning")
	Error("This is a error. ", "This is a error")

	Tracef("This is a %s", "trace")
	Warningf("This is a %s", "warning")
	Errorf("This is a %s", "error")

	SetLevel(LevelWarning)
	Trace("This is a trace. ", "This is a trace")
	Warning("This is a warning. ", "This is a warning")
	Error("This is a error. ", "This is a error")

	Tracef("This is a %s", "trace")
	Warningf("This is a %s", "warning")
	Errorf("This is a %s", "error")

	SetLevel(LevelError)
	Trace("This is a trace. ", "This is a trace")
	Warning("This is a warning. ", "This is a warning")
	Error("This is a error. ", "This is a error")

	Tracef("This is a %s", "trace")
	Warningf("This is a %s", "warning")
	Errorf("This is a %s", "error")
}
