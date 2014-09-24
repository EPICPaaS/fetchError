package main

import (
	"bufio"
	//"fmt"
	"os"
	"path"
	"strings"
	"time"
)

/**
* nginx 日志处理
 */
func nginxHandleFile(filePath string) *FetchRecord {
	baseName := path.Base(filePath)
	newFilePath := filePath
	if !strings.HasSuffix(baseName, ".log") {
		newFilePath = filePath + ".fetch"
		os.Rename(filePath, newFilePath)
	}

	r1, r2 := infrastructureGetFR(newFilePath)
	nginxReadLines(newFilePath, r1, r2)

	if !strings.HasSuffix(baseName, ".log") {
		os.Rename(newFilePath, newFilePath+".fetched."+time.Now().Format("20060102150405"))
	}

	return r2

}

/**
* 读取分析nginx日志
 */
func nginxReadLines(path string, r1, r2 *FetchRecord) error {

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

		// 因nginx错误日志信息都是一行，因此一行里包含里所有错误信息
		if strings.Contains(line, "[error]") || strings.Contains(line, "[crit]") || strings.Contains(line, "[emerg]") {

			//提取该行主要信息
			logRecord.LogLevel, logRecord.LogTime, logRecord.LogClass, logRecord.LogLineNumber, logRecord.LogContent = nginxReadLine(line)
			if r1 != nil && logRecord.LogTime > r1.LastTime {
				lastLog = &logRecord
				writeNginxLogRecord(*lastLog)
			} else {
				continue
			}

		}
	}

	if lastLog != nil {
		//更新日志文件最后分析日期
		r2.LastTime = lastLog.LogTime
	}

	return scanner.Err()
}

/**
* 生成抓取成功的日志信息的数据库数据文件
 */
func writeNginxLogRecord(l LogRecord) {

	if l.LogLevel == "error" || l.LogLevel == "crit" || l.LogLevel == "emerg" {
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

//分析nginx日志关键信息
func nginxReadLine(line string) (logLevel, logTime, logClass, logLineNumber, logContent string) {
	nginxFields := strings.SplitN(line, " ", 7)
	if len(nginxFields) > 6 {
		logTime = nginxFields[0] + " " + nginxFields[1]
		f := nginxFields[2]
		logLevel = f[1 : len(f)-1]
		logClass = nginxFields[5]
		logLineNumber = "0"
		logContent = nginxFields[6]
	} else {
		logLevel = "-"
		logTime = TimeNowString
		logClass = "-"
		logLineNumber = "0"
		logContent = "-"
	}

	return logLevel, logTime, logClass, logLineNumber, logContent
}
