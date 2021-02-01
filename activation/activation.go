package activation

import (
	"encoding/json"
	"errors"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/event"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils/license"
	"github.com/infinit-lab/gravity/utils/machine"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

const (
	TopicActivation string = "activation"
)

const (
	StatusActivated string = "activated"
	StatusUnactivated string = "unactivated"
)

var lic *license.License
var mutex sync.Mutex
var fingerprint string
var status string
var authorizations []license.Authorization
var isForever bool
var duration int
var currentDatetime string

func init() {
	status = StatusUnactivated
	publicPemFile := config.GetString("activation.public")
	if len(publicPemFile) == 0 {
		publicPemFile = "public.pem"
	}
	err := license.SetPublicPem(publicPemFile)
	if err != nil {
		printer.Error(err)
		printer.Error("加载公钥文件失败")
		os.Exit(1)
	}
	fingerprint, err = license.GenerateFingerprint()
	if err != nil {
		printer.Error(err)
		printer.Error("生成机器指纹失败")
		os.Exit(1)
	}
	go func() {
		t := time.Minute
		for {
			e := new(event.Event)
			e.Topic = TopicActivation
			err := updateLocalLicense(int(t / time.Second))
			if err != nil {
				printer.Error(err)
				e.Status = StatusUnactivated
			} else {
				e.Status = StatusActivated
			}
			if status != e.Status {
				status = e.Status
				_ = event.Publish(e)
			}
			time.Sleep(t)
		}
	}()
}

func GetFingerprint() string {
	return fingerprint
}

func getDatetimeDelta() (int, error) {
	if lic == nil {
		return 0, errors.New("未加载授权证书")
	}
	validDateTime, err := time.ParseInLocation("2006-01-02 15:04:05", lic.ValidDatetime, time.UTC)
	if err != nil {
		printer.Error(err)
		return 0, err
	}
	currentDateTime, err := time.ParseInLocation("2006-01-02 15:04:05", lic.CurrentDatetime, time.UTC)
	if err != nil {
		printer.Error(err)
		return 0, err
	}
	if currentDateTime.Before(time.Now()) {
		currentDateTime = time.Now()
	}
	return int(validDateTime.Sub(currentDateTime).Seconds()), nil
}

func getDurationDelta() (int, error) {
	if lic == nil {
		return 0, errors.New("未加载授权文件")
	}
	return lic.ValidDuration, nil
}

func updateDatetime(delta int) error {
	if lic == nil {
		return errors.New("未加载授权文件")
	}
	now := time.Now().UTC()
	original, err  := time.ParseInLocation("2006-01-02 15:04:05", lic.CurrentDatetime, time.UTC)
	if err != nil {
		printer.Error(err)
		return err
	}
	if original.Before(now) {
		original = now
	}
	lic.CurrentDatetime = original.Format("2006-01-02 15:04:05")
	lic.ValidDatetime = original.Add(time.Duration(delta) * time.Second).UTC().Format("2006-01-02 15:04:05")
	return nil
}

func updateDuration(delta int) error {
	if lic == nil {
		return errors.New("未加载授权文件")
	}
	lic.ValidDuration = delta
	lic.CurrentDuration = delta
	return nil
}

func readLicense(licenseFile string) *license.License {
	content, err := ioutil.ReadFile(licenseFile)
	if err != nil {
		printer.Error(err)
		return nil
	}
	data, err := machine.DecodeSelf(string(content))
	if err != nil {
		printer.Error(err)
		return nil
	}
	l := new(license.License)
	err = json.Unmarshal(data, l)
	if err != nil {
		printer.Error(err)
		return nil
	}
	return l
}

func writeLicense(licenseFile string) error {
	if lic == nil {
		printer.Error("未加载授权文件")
	}
	data, err := json.Marshal(lic)
	if err != nil {
		printer.Error(err)
		return err
	}
	content, err := machine.EncodeSelf(data)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = ioutil.WriteFile(licenseFile, []byte(content), os.ModePerm)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func getLocalLicenseFile() string {
	licenseFile := config.GetString("activation.license")
	if len(licenseFile) == 0 {
		licenseFile = "activation.txt"
	}
	return licenseFile
}

func updateLocalLicense(period int) error {
	mutex.Lock()
	defer mutex.Unlock()
	lic = readLicense(getLocalLicenseFile())
	if lic == nil {
		lic = new(license.License)
		lic.IsForever = false
		lic.ValidDatetime = time.Now().UTC().Format("2006-01-02 15:04:05")
		lic.CurrentDatetime = lic.ValidDatetime
		lic.ValidDuration = 0
		lic.CurrentDuration = 0
	}
	dateTimeDelta, err := getDatetimeDelta()
	if err != nil {
		printer.Error(err)
		return err
	}
	durationDelta, err := getDurationDelta()
	if err != nil {
		printer.Error(err)
		return err
	}
	durationDelta -= period
	var delta int
	if dateTimeDelta < durationDelta {
		delta = dateTimeDelta
	} else {
		delta = durationDelta
	}
	if delta < 0 {
		delta = 0
	}
	err = updateDatetime(delta)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = updateDuration(delta)
	if err != nil {
		printer.Error(err)
		return err
	}
	authorizations = lic.Authorizations
	isForever = lic.IsForever
	duration = lic.ValidDuration
	currentDatetime = lic.CurrentDatetime
	err = writeLicense(getLocalLicenseFile())
	if err != nil {
		printer.Error(err)
		return err
	}
	if lic.IsForever == true {
		return nil
	} else if delta > 0 {
		return nil
	}
	return errors.New("授权到期")
}

func GetStatus() string {
	return status
}

func GetAuthorizations() []license.Authorization {
	return authorizations
}

func updateLicense(licenseFile string) error {
	mutex.Lock()
	defer mutex.Unlock()
	content, err := ioutil.ReadFile(licenseFile)
	if err != nil {
		printer.Error(err)
		return err
	}
	l, err := license.LoadLicense(string(content))
	if err != nil {
		printer.Error(err)
		return err
	}
	lic = readLicense(getLocalLicenseFile())
	if lic == nil {
		lic = l
	} else {
		lic.IsForever = l.IsForever
		lic.ValidDuration = l.ValidDuration
		lic.ValidDatetime = l.ValidDatetime
		lic.Authorizations = l.Authorizations
	}
	err = writeLicense(getLocalLicenseFile())
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func UpdateLicense(licenseFile string) error {
	err := updateLicense(licenseFile)
	if err != nil {
		printer.Error(err)
		return err
	}
	e := new(event.Event)
	e.Topic = TopicActivation
	err = updateLocalLicense(0)
	if err != nil {
		printer.Error(err)
		e.Status = StatusUnactivated
	} else {
		e.Status = StatusActivated
	}
	if status != e.Status {
		status = e.Status
		_ = event.Publish(e)
	}
	return nil
}
