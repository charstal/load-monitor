package util

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

func CopyFile(src string, des string) error {
	if len(src) == 0 {
		return errors.New("src empty")
	}

	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(des, input, 0644)
	if err != nil {
		// fmt.Println("Error creating", des)
		// fmt.Println(err)
		return err
	}
	return nil
}

func RenameFile(src string, des string) error {
	if len(src) == 0 {
		return errors.New("src empty")
	}

	if len(des) == 0 {
		return errors.New("des empty")
	}

	err := os.Rename(src, des)
	if err != nil {
		return err
	}
	err = os.Remove(src)
	if err != nil {
		return err
	}

	return nil
}

func GetFileMd5(filename string) (string, error) {
	path := fmt.Sprintf("./%s", filename)
	pFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer pFile.Close()
	md5h := md5.New()
	io.Copy(md5h, pFile)
	return hex.EncodeToString(md5h.Sum(nil)), nil
}
