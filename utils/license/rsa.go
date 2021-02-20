package license

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"io/ioutil"
	"os"
)

func SetPublicPem(publicFile string) error {
	var err error
	publicPem, err = ioutil.ReadFile(publicFile)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func SetPrivatePem(privateFile string) error {
	var err error
	privatePem, err = ioutil.ReadFile(privateFile)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

type reader struct {
}

func (r *reader) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, errors.New("缓存长度为0")
	}
	for i := 0; i < len(buffer); i++ {
		buffer[i] = 0xaa
	}
	return len(buffer), nil
}

func RsaEncrypt(content []byte) ([]byte, error) {
	if publicPem == nil {
		return nil, errors.New("未设置公钥")
	}
	block, _ := pem.Decode(publicPem)
	if block == nil {
		return nil, errors.New("无效公钥")
	}
	publicInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	publicKey, ok := publicInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("非RSA公钥")
	}
	encrypt, err := rsa.EncryptPKCS1v15(&reader{}, publicKey, content)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return encrypt, nil
}

func RsaDecrypt(encrypt []byte) ([]byte, error) {
	if privatePem == nil {
		return nil, errors.New("未设置私钥")
	}
	block, _ := pem.Decode(privatePem)
	if block == nil {
		return nil, errors.New("无效私钥")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	content, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encrypt)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return content, nil
}

func GenerateRsaKey(publicPemPath, privatePemPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 256)
	if err != nil {
		printer.Error(err)
		return err
	}
	derPrivate := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derPrivate,
	}
	privatePem = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPublic, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		printer.Error(err)
		return err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPublic,
	}
	publicPem = pem.EncodeToMemory(block)

	err = ioutil.WriteFile(publicPemPath, publicPem, os.ModePerm)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = ioutil.WriteFile(privatePemPath, privatePem, os.ModePerm)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}
