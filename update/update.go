package update

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/infinit-lab/gravity/printer"
	"io"
	"os"
	"path/filepath"
)

const ext = "update"

var updatePackageName string

func SetUpdatePackageName(t string) {
	updatePackageName = t
}

func ParseUpdatePackage(fileName string) error {
	if len(updatePackageName) == 0 {
		return errors.New("Invalid device type. ")
	}
	archive, err := zip.OpenReader(fileName)
	if err != nil {
		printer.Error(err)
		return err
	}
	defer func() {
		_ = archive.Close()
		_ = os.Remove(fileName)
	}()

	isFind := false
	for _, f := range archive.File {
		if fmt.Sprintf("%s.%s", updatePackageName, ext) != f.Name {
			continue
		}
		err = unzip(f)
		if err != nil {
			printer.Error(err)
			return err
		}
		isFind = true
	}
	if !isFind {
		return errors.New("No update package. ")
	}
	return nil
}

func Update() error {
	if len(updatePackageName) == 0 {
		return errors.New("Invalid device type. ")
	}
	fileName := fmt.Sprintf("%s.%s", updatePackageName, ext)
	archive, err := zip.OpenReader(fileName)
	if err != nil {
		printer.Error(err)
		return err
	}
	defer func() {
		_ = archive.Close()
		_ = os.Remove(fileName)
	}()
	for _, f := range archive.File {
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(f.Name, os.ModePerm)
			if err != nil {
				printer.Error(err)
			}
			continue
		}
		err = os.MkdirAll(filepath.Dir(f.Name), os.ModePerm)
		if err != nil {
			printer.Error(err)
			continue
		}
		err = unzip(f)
		if err != nil {
			printer.Error(err)
			continue
		}
	}
	return nil
}

func unzip(f *zip.File) error {
	var dst *os.File
	var fileInArchive io.ReadCloser
	var err error
	defer func() {
		if dst != nil {
			_ = dst.Close()
			dst = nil
		}
		if fileInArchive != nil {
			_ = fileInArchive.Close()
			fileInArchive = nil
		}
	}()

	dst, err = os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		printer.Error(err)
		return err
	}
	fileInArchive, err = f.Open()
	if err != nil {
		printer.Error(err)
		return err
	}
	_, err = io.Copy(dst, fileInArchive)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}
