package main

import (
	"flag"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils/license"
	"io/ioutil"
	"os"
	"time"
)

var isForever bool
var datetime string
var fingerprint string

func main() {
	flag.StringVar(&fingerprint, "p", "", "机器指纹(必填）")
	flag.BoolVar(&isForever, "f", false, "是否永久")
	flag.StringVar(&datetime, "d",
		time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05"), "到期时间")
	flag.Parse()
	if len(fingerprint) == 0 {
		flag.Usage()
		return
	}
	validDatetime, err := time.ParseInLocation("2006-01-02 15:04:05", datetime, time.Local)
	if err != nil {
		printer.Error(err)
		return
	}
	if validDatetime.Before(time.Now()) {
		printer.Error("无效到期时间")
		return
	}
	privatePemFile := config.GetString("register.private")
	if len(privatePemFile) == 0 {
		privatePemFile = "private.pem"
	}
	err = license.SetPrivatePem(privatePemFile)
	if err != nil {
		printer.Error(err)
		return
	}
	var lic license.License
	lic.IsForever = isForever
	lic.ValidDatetime = validDatetime.UTC().Format("2006-01-02 15:04:05")
	lic.CurrentDatetime = time.Now().UTC().Format("2006-01-02 15:04:05")
	lic.ValidDuration = int(validDatetime.Sub(time.Now()).Seconds())
	lic.CurrentDuration = lic.ValidDuration
	content, err := license.GenerateLicense(fingerprint, lic)
	if err != nil {
		printer.Error(err)
		return
	}
	err = ioutil.WriteFile(fingerprint + ".cert", []byte(content), os.ModePerm)
	if err != nil {
		printer.Error(err)
		return
	}
}
