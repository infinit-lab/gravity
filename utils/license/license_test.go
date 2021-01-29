package license

import (
	"github.com/infinit-lab/gravity/printer"
	"testing"
)

func TestGenerateLicense(t *testing.T) {
	err := GenerateRsaKey("public.pem", "private.pem")
	if err != nil {
		t.Fatal(err)
	}
	err = SetPublicPem("public.pem")
	if err != nil {
		t.Fatal(err)
	}
	err = SetPrivatePem("private.pem")
	if err != nil {
		t.Fatal(err)
	}
	license := License{
		IsForever:      false,
		ValidDatetime:  "2021-01-27 00:00:00",
		ValidDuration:  3567,
		Authorizations: nil,
	}
	fingerprint, err := GenerateFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	printer.Trace(fingerprint)
	printer.Trace(len(fingerprint))
	content, err := GenerateLicense(fingerprint, license)
	if err != nil {
		t.Fatal(err)
	}
	lic, err := LoadLicense(content)
	if err != nil {
		t.Fatal(err)
	}
	printer.Trace(lic)
}
