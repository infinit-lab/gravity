package license

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils/machine"
)

type Authorization struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	ValueType string   `json:"valueType"`
	Values    []string `json:"values"`
	Current   string   `json:"current,omitempty"`
}

type License struct {
	IsForever       bool            `json:"isForever"`
	ValidDatetime   string          `json:"validDatetime"`
	CurrentDatetime string          `json:"currentDatetime"`
	ValidDuration   int             `json:"validDuration"`
	CurrentDuration int             `json:"currentDuration"`
	Authorizations  []Authorization `json:"authorizations,omitempty"`
}

var publicPem []byte
var privatePem []byte

func GenerateFingerprint() (string, error) {
	if publicPem == nil {
		return "", errors.New("未设置公钥")
	}
	f, err := machine.GetMachineFingerprint()
	if err != nil {
		printer.Error(err)
		return "", err
	}
	fingerprint, err := RsaEncrypt(f)
	if err != nil {
		printer.Error(err)
		return "", err
	}
	return hex.EncodeToString(fingerprint), nil
}

func GenerateLicense(fingerprint string, license License) (string, error) {
	if privatePem == nil {
		return "", errors.New("未设置私钥")
	}
	f, err := hex.DecodeString(fingerprint)
	if err != nil {
		printer.Error(err)
		return "", err
	}
	f, err = RsaDecrypt(f)
	if err != nil {
		printer.Error(err)
		return "", err
	}
	data, err := json.Marshal(license)
	if err != nil {
		printer.Error(err)
		return "", err
	}
	return machine.Encode(f, data)
}

func LoadLicense(lic string) (*License, error) {
	fingerprint, err := machine.GetMachineFingerprint()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	encrypt, err := machine.Decode(fingerprint, lic)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	license := new(License)
	err = json.Unmarshal(encrypt, license)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return license, nil
}
