package license

import (
	"github.com/infinit-lab/gravity/printer"
	"testing"
)

func TestGenerateRsaKey(t *testing.T) {
	err := GenerateRsaKey("public.pem", "private.pem")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRsaDecrypt(t *testing.T) {
	err := SetPublicPem("public.pem")
	if err != nil {
		t.Fatal(err)
	}
	err = SetPrivatePem("private.pem")
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("TestRsaDecrypt11")
	encrypt, err := RsaEncrypt(content)
	if err != nil {
		t.Fatal(err)
	}
	data, err := RsaDecrypt(encrypt)
	if err != nil {
		t.Fatal(err)
	}
	printer.Trace(string(data))
}
