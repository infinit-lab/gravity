package machine

import (
	"log"
	"testing"
)

func TestGetMachineFingerprint(t *testing.T) {
	fingerprint, err := GetMachineFingerprint()
	if err != nil {
		t.Fatal("Failed to GetMachineFingerprint. error: ", err)
	}
	log.Println("Machine Fingerprint is ", fingerprint)

	testStr := []byte("Hello World")
	encodeData, err := Encode(fingerprint, testStr)
	if err != nil {
		t.Fatal("Failed to Encode. error: ", err)
	}
	log.Println("Encode: ", encodeData)

	decodeData, err := Decode(fingerprint, encodeData)
	if err != nil {
		t.Fatal("Failed to Decode. error: ", err)
	}
	log.Println("Decode: ", string(decodeData))

	encodeData, err = EncodeSelf(testStr)
	if err != nil {
		t.Fatal("Failed to EncodeSelf. error: ", err)
	}
	log.Println("Encode: ", encodeData)

	decodeData, err = DecodeSelf(encodeData)
	if err != nil {
		t.Fatal("Failed to DecodeSelf. error: ", err)
	}
	log.Println("Decode: ", string(decodeData))
}
