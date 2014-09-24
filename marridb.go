package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

func marridbHandleFile(filePath string) *FetchRecord {

	baseName := path.Base(filePath)
	newFilePath := filePath
	if !strings.HasSuffix(baseName, ".log") {
		newFilePath = filePath + ".fetch"
		os.Rename(filePath, newFilePath)
	}

	//获取该文件分析记录
	r1, r2 := infrastructureGetFR(newFilePath)
	marridbReadLines(newFilePath, r1, r2)

	//标记文件分析完毕
	if !strings.HasSuffix(baseName, ".log") {
		os.Rename(newFilePath, newFilePath+".fetched."+time.Now().Format("20060102150405"))
	}
	return r2
}

func marridbReadLines(path string, r1, r2 *FetchRecord) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	logRecord := LogRecord{r2.IP, r2.AppID, r2.ModuleName, r2.Port, "", "", "", "", "", false, "E", path}
	var lastLog *LogRecord

	for scanner.Scan() {
		line := scanner.Text()
		line2 := strings.TrimSpace(line)
		if len(line2) == 0 { //空行跳出当前循环
			continue
		}

		if strings.Contains(line, "[ERROR]") || strings.Contains(line, "[Note]") || strings.Contains(line, "[Warning]") {
			if lastLog == nil {
				//提取该行主要信息
				logRecord.LogLevel, logRecord.LogTime, logRecord.LogClass, logRecord.LogLineNumber, logRecord.LogContent = marridbReadLine(line)

				if r1 != nil && logRecord.LogTime > r1.LastTime {
					lastLog = &logRecord
				} else {
					continue
				}

			} else {
				//处理上一条
				if lastLog != nil {
					writeMarridbLogRecord(*lastLog)
					lastLog = nil
				}
				//处理当前这条
				logLevel, logTime, logClass, logLineNumber, logContent := marridbReadLine(line)

				if r1 != nil && logTime > r1.LastTime {
					logRecord := LogRecord{r2.IP, r2.AppID, r2.ModuleName, r2.Port, logLevel, logTime, logClass, logLineNumber, logContent, false, "E", path}
					lastLog = &logRecord
				} else {
					continue
				}
			}

		} else if lastLog != nil { //拼接详细信息
			lastLog.LogContent = lastLog.LogContent + "\n" + line
		}

	}

	if lastLog != nil {
		//处理做后一条
		r2.LastTime = lastLog.LogTime //更新日志文件最后分析日期
		writeMarridbLogRecord(*lastLog)
	}

	return scanner.Err()
}

//分析nginx日志关键信息
func marridbReadLine(line string) (logLevel, logTime, logClass, logLineNumber, logContent string) {

	line = strings.Replace(line, "  ", " ", -1)
	files := strings.SplitAfterN(line, " ", 4)
	if len(files) > 3 {
		f := strings.TrimSpace(files[2])
		logLevel = f[1 : len(f)-1]
		logTime = "20" + files[0] + " " + files[1]
		logTimeDate, _ := time.Parse("20060102 15:04:05", strings.TrimSpace(logTime))
		logTime = logTimeDate.Format("2006/01/02 15:04:05")
		fmt.Println(logTime)
		logClass = "marridb"
		logLineNumber = "0"
		logContent = files[3]
	} else {
		logLevel = "-"
		logTime = TimeNowString
		logClass = "-"
		logLineNumber = "0"
		logContent = "-"
	}
	return logLevel, logTime, logClass, logLineNumber, logContent
}

/**
* 生成抓取成功的日志信息的数据库数据文件
 */
func writeMarridbLogRecord(l LogRecord) {
	if l.LogLevel == "ERROR" {
		c := l.LogContent + "\n"
		line := strings.Join([]string{l.LogTime, l.LogClass, l.LogLineNumber, l.LogException, l.IP, l.Port, l.ModuleName, l.AppID, c, l.LogFilePath}, "\x1F")
		BufWriter.WriteString(line)
		BufWriter.WriteString("\x1E")

		writeIterNum++
		if writeIterNum > 1024 {
			BufWriter.Flush()
			writeIterNum = 1
		}
	}
}
