package update

import "testing"

func TestSetDeviceType(t *testing.T) {
	SetUpdatePackageName("stairwell")
}

func TestParseUpdatePackage(t *testing.T) {
	err := ParseUpdatePackage("stairwell_package." + ext)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdate(t *testing.T) {
	err := Update()
	if err != nil {
		t.Fatal(err)
	}
}
