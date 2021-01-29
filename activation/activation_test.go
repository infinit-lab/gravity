package activation

import (
	"encoding/hex"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils/license"
	"os"
	"testing"
	"time"
)

func TestGetFingerprint(t *testing.T) {
	os.Args = append(os.Args, "rsa.dynamic=true")
	config.LoadArgs()

	fingerprint := GetFingerprint()
	printer.Trace(fingerprint)

	err := license.SetPrivatePem("private.pem")
	if err != nil {
		t.Fatal(err)
	}
	data, err := hex.DecodeString(fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	f, err := license.RsaDecrypt(data)
	if err != nil {
		t.Fatal(err)
	}
	printer.Trace(f)

	time.Sleep(time.Second)

	printer.Trace(GetStatus())
}
