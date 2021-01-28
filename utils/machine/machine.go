package machine

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils/baseboard"
	"github.com/infinit-lab/gravity/utils/cpu"
	"github.com/infinit-lab/gravity/utils/harddisk"
)

func GetMachineFingerprint() ([]byte, error) {
	baseboardUuid, err := baseboard.GetUUID()
	if err != nil {
		printer.Error("Failed to GetBaseBoardUUID. error: ", err)
		return nil, err
	}
	cpuId, err := cpu.GetCpuID()
	if err != nil {
		printer.Error("Failed to GetCpuID. error: ", err)
		return nil, err
	}
	diskIds, err := harddisk.GetDiskSerialNumber()
	if err != nil {
		printer.Error("Failed to GetDiskSerialNumber. error: ", err)
		return nil, err
	}
	if len(diskIds) == 0 {
		return nil, errors.New("未获取到磁盘信息")
	}
	str := baseboardUuid + cpuId + diskIds[0]
	encode := base64.StdEncoding.EncodeToString([]byte(str))
	has := md5.Sum([]byte(encode))
	return has[:], nil
}

func generateKey(fingerprint []byte) ([]byte, error) {
	has := md5.Sum(fingerprint)
	return has[:], nil
}

func Encode(fingerprint []byte, content []byte) (string, error) {
	key, err := generateKey(fingerprint)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		printer.Error("Failed to NewCipher. error: ", err)
		return "", err
	}
	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)
	dst := make([]byte, len(content))
	stream.XORKeyStream(dst, content)
	return hex.EncodeToString(dst), nil
}

func EncodeSelf(content []byte) (string, error) {
	key, err := GetMachineFingerprint()
	if err != nil {
		return "", err
	}
	return Encode(key, content)
}

func Decode(fingerprint []byte, encryptData string) ([]byte, error) {
	data, err := hex.DecodeString(encryptData)
	if err != nil {
		return nil, err
	}
	key, err := generateKey(fingerprint)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		printer.Error("Failed to NewCipher. error: ", err)
		return nil, err
	}
	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)
	dst := make([]byte, len(data))
	stream.XORKeyStream(dst, data)
	return dst, nil
}

func DecodeSelf(encryptData string) ([]byte, error) {
	key, err := GetMachineFingerprint()
	if err != nil {
		return nil, err
	}
	return Decode(key, encryptData)
}
