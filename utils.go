package main

import (
	//"bufio"
	"errors"
	//"fmt"
	//"log"
	"net"
	"os"
	"path"
	"path/filepath"
	//"runtime"
	"strings"
	//"time"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func LocalIP() (net.IP, error) {
	tt, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 { // loopback address
				continue
			}
			return v4, nil
		}
	}
	return nil, errors.New("cannot find local IP address")
}

func WriteFile(filePath, fileContent string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	CheckErr(err)
	defer file.Close()

	file.WriteString(fileContent)
}

func GetFilelist(path1 string) []string {
	var all []string
	filepath.Walk(path1, func(path2 string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		baseName := path.Base(path2)
		if strings.HasPrefix(baseName, "tmp.") || strings.Contains(baseName, ".fetch") || strings.Contains(path2, "infrastructure/") {

		} else if strings.Contains(baseName, ".log") {
			all = append(all, path2)
		}
		return nil
	})

	return all
}
