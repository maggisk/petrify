package main

import "os"
import "io"
import "path/filepath"
import "strings"

func CopyDir(src string, dst string) {
	checkError(os.MkdirAll(dst, 0755))

	var visit func(string, os.FileInfo, error) error
	visit = func(path string, fileInfo os.FileInfo, err error) error {
		checkError(err)
		thisDst := strings.Replace(path, src, dst, 1)
		if fileInfo.IsDir() {
			checkError(os.MkdirAll(thisDst, 0755))
		} else {
			checkError(CopyFile(path, thisDst))
		}
		return nil
	}
	filepath.Walk(src, visit)
}

func CopyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
