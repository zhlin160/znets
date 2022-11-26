package znets

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

type log struct {
	TimeStamp string `json:"timestamp"`
	Level     string `json:"level"`
	FileName  string `json:"file_name"`
	Content   string `json:"content"`
}

type HLog struct {
	workPath    string
	workPathLen int
	model       string
}

const (
	RED = uint8(iota + 91)
	GREEN
	YELLOW
	BLUE
	MAGENTA

	INFO    = "[INFO]"
	TRAC    = "[TRAC]"
	ERROR   = "[ERROR]"
	WARN    = "[WARNING]"
	SUCCESS = "[SUCCESS]"
)

func NewLog() *HLog {
	return buildLog("")
}

func NewLogWithModel(model string) *HLog {
	return buildLog(model)
}

func buildLog(model string) *HLog {
	path, err := os.Getwd()
	if err != nil {
		path = ""
	}
	pathLen := len(path)
	if model == "" {
		model = "dev"
	}
	return &HLog{
		workPath:    path,
		workPathLen: pathLen,
		model:       model,
	}
}

//返回执行log的文件名称及行号
func (l *HLog) fileName() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = ""
		line = 0
	}
	if file != "" {
		file = file[l.workPathLen+1:]
	}
	return file + ":" + strconv.Itoa(line)
}

func (l *HLog) Info(format string, args ...interface{}) {
	l.showLog(blue(INFO), format, args...)
}

func (l *HLog) Trace(format string, args ...interface{}) {
	l.showLog(yellow(TRAC), format, args...)
}

func (l *HLog) Error(format string, args ...interface{}) {
	l.showLog(red(ERROR), format, args...)
}

func (l *HLog) Warning(format string, args ...interface{}) {
	l.showLog(magenta(WARN), format, args...)
}

func (l *HLog) Success(format string, args ...interface{}) {
	l.showLog(green(SUCCESS), format, args...)
}

func (l *HLog) showLog(prefix, format string, args ...interface{}) {
	fileName := l.fileName()
	content := fmt.Sprintf(format, args...)
	currTime := formatTimeByCurrent()
	fmt.Println(fmt.Sprintf("%s %s %s %s\n", prefix, currTime, blue(fileName), content))
	if l.model == "production" {
		l.writeFile(prefix, currTime, fileName, content)
	}
}

func (l *HLog) writeFile(level, timestamp, fileName, content string) {
	log := log{
		TimeStamp: timestamp,
		Level:     level,
		FileName:  fileName,
		Content:   content,
	}
	buffer, _ := json.Marshal(&log)
	logDir := l.workPath + "/log/"
	if _, err := os.Stat(logDir); err != nil && os.IsNotExist(err) {
		err = os.Mkdir(logDir, os.ModePerm)
		if err != nil {
			fmt.Printf("日志文件夹创建失败, err=%s", err.Error())
			return
		}
	}
	fl, err := os.OpenFile(logDir+time.Now().Format("20060102")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("日志文件创建失败, err=%s", err.Error())
		return
	}
	buffer = append(buffer, '\r', '\n', '\r', '\n')
	defer fl.Close()
	fl.Write(buffer)
}

func red(content string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", RED, content)
}

func green(content string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", GREEN, content)
}

func yellow(content string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", YELLOW, content)
}

func blue(content string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", BLUE, content)
}

func magenta(content string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", MAGENTA, content)
}

func formatTimeByCurrent() string {
	return time.Now().Format(time.RFC3339)
}
