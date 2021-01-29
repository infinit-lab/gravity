package main

import (
	"fmt"
	"github.com/infinit-lab/gravity/config"
	"github.com/infinit-lab/gravity/printer"
	"github.com/infinit-lab/gravity/utils/license"
)

func main() {
	publicPemFile := config.GetString("fingerprint.public")
	if len(publicPemFile) == 0 {
		publicPemFile = "public.pem"
	}
	err := license.SetPublicPem(publicPemFile)
	if err != nil {
		printer.Error(err)
		return
	}
	f, err := license.GenerateFingerprint()
	if err != nil {
		printer.Error(err)
		return
	}
	fmt.Println(f)
}
