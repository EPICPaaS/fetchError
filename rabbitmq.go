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
func rabbimqHandleFile(filePath string) *FetchRecord {
	baseName := path.Base(filePath)
	newFilePath := filePath

	if !strings.HasSuffix(baseName, ".log") {
		newFilePath = filePath + ".fetch"
		os.Rename(filePath, newFilePath)
	}

	r1, r2 := infrastructureGetFR(newFilePath)
	rabbitmqReadLines(newFilePath, r1, r2)

	if !strings.HasSuffix(baseName, ".log") {
		os.Rename(newFilePath, newFilePath+".fetched."+time.Now().Format("20060102150405"))
	}

	return r2
}

/**
* 读取分析rabbitmq日志
 */
func rabbitmqReadLines(path string, r1, r2 *FetchRecord) error {
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

		if strings.HasPrefix(line, "=ERROR REPORT") {

			if lastLog == nil {
				logRecord.LogLevel, logRecord.LogTime, logRecord.LogClass, logRecord.LogLineNumber, logRecord.LogContent = rabbitmqReadLine(line)
				if r1 != nil && logRecord.LogTime > r1.LastTime {
					lastLog = &logRecord

				} else {
					continue
				}

			}

		} else if lastLog != nil {

			lastLog.LogContent = lastLog.LogContent + "\n" + line
			//当出现空行时该条异常错误信息结束
			if len(strings.TrimSpace(line)) == 0 {
				//处理
				r2.LastTime = lastLog.LogTime //更新最后时间
				writeRabbitmqLogRecord(*lastLog)
				lastLog = nil
				continue
			}

		}
	}

	return scanner.Err()
}

//读取异常信息
func rabbitmqReadLine(line string) (logLevel, logTime, logClass, logLineNumber, logContent string) {

	fields := strings.SplitN(line, " ", 4)
	if len(fields) > 3 {
		logTime = fields[2]
		logTimeDate, err := time.Parse("2-Jan-2006::15:04:05", logTime)
		if err == nil {
			logTime = logTimeDate.Format("2006/01/02 15:04:05")
		}
		tf := fields[0]
		logLevel = tf[1:len(tf)]
		logClass = "rabbitmq"
		logLineNumber = "0"
		logContent = fields[3]
	} else {
		logLevel = "-"
		logTime = TimeNowString
		logClass = "rabbitmq"
		logLineNumber = "0"
		logContent = "-"
	}

	return logLevel, logTime, logClass, logLineNumber, logContent
}

/**
* 生成抓取成功的日志信息的数据库数据文件
 */
func writeRabbitmqLogRecord(l LogRecord) {

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
