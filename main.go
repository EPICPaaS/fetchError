// fetchError project main.go
package main

import (
	"bufio"
	//"errors"
	"fmt"
	"log"
	"os"
	"path"
	//"path/filepath"
	"flag"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"
)

const BufSize = 4096 * 2

var TIME_NOW time.Time
var LOCAL_IP string = "127.0.0.1"
var writeIterNum int = 1

//var file_path string = "/home/yourchanges/paas_home_dev/logs/services/taskscheduler-1.0.0/taskscheduler_srv/taskscheduler_srv_10000.log.1"
var FetchRecords []FetchRecord
var BufWriter *bufio.Writer
var (
	AppPath       string
	AppConfigPath string
	OutPath       string
	TimeNowString string
)

type FetchRecord struct {
	IP         string
	AppID      string
	ModuleName string
	Port       string
	LastTime   string
}

type LogRecord struct {
	IP            string
	AppID         string
	ModuleName    string
	Port          string
	LogLevel      string
	LogTime       string
	LogClass      string
	LogLineNumber string
	LogContent    string
	HasMulti      bool
	LogException  string
	LogFilePath   string
}

func init() {
	os.Chdir(path.Dir(os.Args[0]))
	AppPath = path.Dir(os.Args[0])
	AppConfigPath = path.Join(AppPath, "fetch_record")
	OutPath = path.Join(AppPath, "out")
	//fmt.Println(AppConfigPath)
	FetchRecords = readContent()
}

func WriteConfigFile(fileContent string) {
	file, err := os.OpenFile(AppConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	CheckErr(err)
	defer file.Close()

	file.WriteString(fileContent)
}

func readContent() []FetchRecord {
	file, err := os.Open(AppConfigPath)
	var records []FetchRecord
	if err != nil {
		return records
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) > 4 {
			record := FetchRecord{fields[0], fields[1], fields[2], fields[3], fields[4]}
			records = append(records, record)
		}

	}
	return records
}

func getFetchRecord(ip, appID, moduleName, port string) *FetchRecord {
	for _, rr := range FetchRecords {
		if rr.IP == ip && rr.AppID == appID && rr.ModuleName == moduleName && rr.Port == port {
			return &rr
		}
	}
	return &FetchRecord{ip, appID, moduleName, port, "0000/00/00 00:00:00"}
}

func insertOrUpdate(r FetchRecord) {
	hasOne := false
	for i, rr := range FetchRecords {
		//update
		if rr.IP == r.IP && rr.AppID == r.AppID && rr.ModuleName == r.ModuleName && rr.Port == r.Port {
			if rr.LastTime < r.LastTime {
				var frs []FetchRecord
				frs = FetchRecords[0:i]
				frs = append(frs, r)
				frs = append(frs, FetchRecords[i+1:]...)
				FetchRecords = frs
			}
			return

		}
	}
	//insert
	if !hasOne {
		FetchRecords = append(FetchRecords, r)
	}
}

func writeFetchRecords() {
	outputFile, err := os.OpenFile(AppConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	CheckErr(err)
	defer outputFile.Close()

	w := bufio.NewWriter(outputFile)
	for _, rr := range FetchRecords {
		//line := rr.IP + "," + rr.AppID + "," + rr.ModuleName + "," + rr.Port + "," + rr.LastTime
		line := strings.Join([]string{rr.IP, rr.AppID, rr.ModuleName, rr.Port, rr.LastTime}, ",")
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func writeLogRecord(l LogRecord) {
	if l.LogLevel == "ERROR" || strings.Count(l.LogContent, "Exception") > 0 || strings.Count(l.LogContent, ".OutOfMemoryError") > 0 {
		//line := strings.Join([]string{l.IP, l.AppID, l.ModuleName, l.Port, l.LogLevel, l.LogTime, l.LogClass, l.LogLineNumber, l.LogContent}, ",")
		c := l.LogContent + "\n"
		//c = strings.Replace(c, "\t", "  ", -1)
		if strings.Count(l.LogContent, ".OutOfMemoryError") > 0 {
			l.LogException = "java.lang.OutOfMemoryError"
		}
		line := strings.Join([]string{l.LogTime, l.LogClass, l.LogLineNumber, l.LogException, l.IP, l.Port, l.ModuleName, l.AppID, c, l.LogFilePath}, "\x1F")
		//line := l.IP + "," + l.AppID + "," + l.ModuleName + "," + l.Port + "," + l.LogLevel + "," + l.LogTime + "," + l.LogClass + "," + l.LogLineNumber + "," + l.LogContent
		//fmt.Fprintln(BufWriter, line)
		//BufWriter.WriteString("\x1F")
		BufWriter.WriteString(line)
		BufWriter.WriteString("\x1E")
		//fmt.Fprintf(BufWriter, "%s,%s,%s,%s,%s,%s,%s,%s,%s\n", l.IP, l.AppID, l.ModuleName, l.Port, l.LogLevel, l.LogTime, l.LogClass, l.LogLineNumber, l.LogContent)
		writeIterNum = writeIterNum + 1
		if writeIterNum > 1024 {
			BufWriter.Flush()
			writeIterNum = 1
		}

	}
}

func GetFetchRecord(filePath string) (*FetchRecord, *FetchRecord) {
	filename := path.Base(filePath)
	fns := strings.Split(filename, ".")
	flevels := strings.Split(filePath, "/")
	size := len(flevels)
	if size < 3 || len(fns) < 2 {
		return nil, nil
	}
	moduleName := flevels[size-2]
	appID := flevels[size-3]

	ns := fns[0]
	i := strings.LastIndex(ns, "_")
	port := ns[i+1:]
	return getFetchRecord(LOCAL_IP, appID, moduleName, port), &FetchRecord{LOCAL_IP, appID, moduleName, port, "0000/00/00 00:00:00"}

}

func handleFile(filePath string) *FetchRecord {
	baseName := path.Base(filePath)
	newFilePath := filePath
	if !strings.HasSuffix(baseName, ".log") {
		newFilePath = filePath + ".fetch"
		os.Rename(filePath, newFilePath)
	}
	r1, r2 := GetFetchRecord(newFilePath)
	readLines(newFilePath, r1, r2)

	if !strings.HasSuffix(baseName, ".log") {
		os.Rename(newFilePath, newFilePath+".fetched."+time.Now().Format("20060102150405"))
	}

	return r2
}

func readLines(path string, r1, r2 *FetchRecord) error {
	file, err := os.Open(path)
	//log.Println(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	logRecord := LogRecord{r2.IP, r2.AppID, r2.ModuleName, r2.Port, "", "", "", "", "", false, "E", path}
	var lastLog *LogRecord
	//i := 1
	//j := 1
	for scanner.Scan() {
		line := scanner.Text()
		line2 := strings.TrimSpace(line)
		if len(line2) == 0 {
			continue
		}
		//j = j + 1

		if strings.HasPrefix(line, "[ERROR") || strings.HasPrefix(line, "[WARN") || strings.HasPrefix(line, "[INFO") || strings.HasPrefix(line, "[DEBUG") {
			if lastLog == nil {
				logRecord.LogLevel, logRecord.LogTime, logRecord.LogClass, logRecord.LogLineNumber, logRecord.LogContent = readLine(line)
				if r1 != nil && logRecord.LogTime > r1.LastTime {
					lastLog = &logRecord
				} else {
					continue
				}

			} else {
				//处理上一条
				if lastLog != nil {
					writeLogRecord(*lastLog)
					//i = i + 1
					lastLog = nil
				}

				//处理当前这条
				logLevel, logTime, logClass, logLineNumber, logContent := readLine(line)

				if r1 != nil && logTime > r1.LastTime {
					logRecord := LogRecord{r2.IP, r2.AppID, r2.ModuleName, r2.Port, logLevel, logTime, logClass, logLineNumber, logContent, false, "E", path}
					lastLog = &logRecord
				} else {
					continue
				}
			}

		} else {
			if lastLog != nil {
				//如果是exception的第二行
				if lastLog.HasMulti == false {
					if strings.Count(line, "Exception:") > 0 {
						idx := strings.Index(line, ":")
						lastLog.LogException = line[0:idx]
						if !strings.HasSuffix(lastLog.LogException, "Exception") {
							lastLog.LogException = "E"
						}
					} else {
						lastLog.LogException = "E"
					}
				}

				lastLog.HasMulti = true
				lastLog.LogContent = lastLog.LogContent + "\n" + line
			}
		}
	}
	if lastLog != nil {
		//处理上一条
		r2.LastTime = lastLog.LogTime
		writeLogRecord(*lastLog)
	}
	//log.Println(i, j)

	return scanner.Err()
}

func readLine(line string) (logLevel, logTime, logClass, logLineNumber, logContent string) {
	fields := strings.SplitN(line, "-", 4)
	//log.Println(fields[3])
	/**
	for i, f := range fields {
		if i == 0 {
			logLevel = f[1 : len(f)-1]
			logLevel = strings.TrimSpace(logLevel)

		} else if i == 1 {
			logTime = f[1 : len(f)-1]
		} else if i == 2 {
			ff := f[1 : len(f)-1]
			ffs := strings.Split(ff, ":")
			if len(ffs) > 1 {
				logClass, logLineNumber = ffs[0], ffs[1]
			}
		} else {
			//if len(logContent) == 0 {
			logContent = f
			//} else {
			//	logContent = logContent + "-" + f
			//}

		}
	}
	*/
	if len(fields) > 3 {
		f := fields[0]
		logLevel = f[1 : len(f)-1]
		logLevel = strings.TrimRight(logLevel, " ")

		f = fields[1]
		logTime = f[1 : len(f)-1]

		f = fields[2]
		ff := f[1 : len(f)-1]
		ffs := strings.Split(ff, ":")
		if len(ffs) > 1 {
			logClass, logLineNumber = ffs[0], ffs[1]
		}

		logContent = fields[3]
	} else {
		logLevel = "-"
		logTime = TimeNowString
		logClass = "-"
		logLineNumber = "0"
		logContent = "-"
	}

	return logLevel, logTime, logClass, logLineNumber, logContent
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ddir := "/home/paas/paas/logs"
	//ddir := "/home/yourchanges/paas_home_dev/logs"

	cpuprofile := flag.String("cpuprofile", "", "the cpu profile ")
	dir := flag.String("dir", "", "the paas logs dir")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		CheckErr(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *dir != "" {
		ddir = *dir
	}

	ip, err := LocalIP()
	CheckErr(err)
	LOCAL_IP = ip.String()
	TIME_NOW = time.Now()
	TimeNowString = "1900/01/01 00:00:00"

	outputFile, err := os.OpenFile(OutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	CheckErr(err)
	defer outputFile.Close()

	BufWriter = bufio.NewWriterSize(outputFile, BufSize)

	all := GetFilelist(ddir)
	var allUpdateRecord []FetchRecord
	for _, fp := range all {
		fr := handleFile(fp)
		allUpdateRecord = append(allUpdateRecord, *fr)
	}

	//最终刷新
	BufWriter.Flush()

	for _, fr := range allUpdateRecord {
		insertOrUpdate(fr)
	}
	writeFetchRecords()
	log.Println("Finished")
}
