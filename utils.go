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

		//读取链接目录下文件
		if (f.Mode()&os.ModeSymlink) > 0 && strings.Contains(path2, "infrastructure/") { //只遍历infrastructure下的软连接目录/排除syslog链接
			linkPath, err := filepath.EvalSymlinks(path2)
			if err != nil {
				return err
			}
			tmpall := GetFilelist(linkPath)
			all = append(all, tmpall[0:]...)
			return nil
		}

		baseName := path.Base(path2)
		if strings.HasPrefix(baseName, "tmp.") || strings.Contains(baseName, ".fetch") {

		} else if strings.Contains(baseName, ".log") {
			all = append(all, path2)
		}
		return nil
	})
	return all
}

//删除string类型slice中的空格
func RemoveSliceSpaceStr(s []string) []string {
	for i, v := range s {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			var tmpSlice []string
			tmpSlice = s[0:i]
			tmpSlice = append(tmpSlice, s[i+1:]...)
			s = tmpSlice
		}
	}
	return s
}
