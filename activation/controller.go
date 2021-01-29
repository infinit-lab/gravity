package activation

import (
	"github.com/gin-gonic/gin"
	"github.com/infinit-lab/gravity/controller"
	"github.com/infinit-lab/gravity/printer"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"strings"
)

type Activation struct {
	Fingerprint string `json:"fingerprint"`
	IsForever   bool   `json:"isForever"`
	Duration    int    `json:"duration"`
}

func init() {
	ctrl, _ := controller.New("/api/1/activation")
	ctrl.GET("", func(context *gin.Context, session *controller.Session) (interface{}, error) {
		a := Activation{
			Fingerprint: fingerprint,
			IsForever:   isForever,
			Duration:    duration,
		}
		return a, nil
	})
	ctrl.PUT("", func(context *gin.Context, session *controller.Session) (interface{}, error) {
		data, err := context.GetRawData()
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		f := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
		err = ioutil.WriteFile(f, data, os.ModePerm)
		if err != nil {
			printer.Error(err)
			return nil, err
		}
		err = UpdateLicense(f)
		if err != nil {
			printer.Error(err)
		}
		_ = os.Remove(f)
		return nil, err
	})
}
