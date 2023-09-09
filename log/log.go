package log

import (
	"fmt"
	"github.com/Ericwyn/GoTools/date"
	"sync"
	"time"
)

type Level int

const (
	LevelError Level = 1
	LevelInfo  Level = 2
	LevelDebug Level = 3
)

const timeFormat = "[MMdd-HHmmss]"

type logTag string

const (
	info  logTag = "INFO"
	debug logTag = "DBUG"
	err   logTag = "ERR "
)

var defaultLevel = LevelInfo
var defaultLogName = "GLOG"

func Init(logName string) {
	defaultLogName = logName
}

func InitWithLevel(logName string, logLevel Level) {
	defaultLogName = logName
	defaultLevel = logLevel
}

func SetLogLevel(logLevel Level) {
	defaultLevel = logLevel
}

func GetLogLevel() Level {
	return defaultLevel
}

func E(msg ...interface{}) {
	if defaultLevel >= LevelError {
		printLog(err, msg...)
	}
}

func I(msg ...interface{}) {
	if defaultLevel >= LevelInfo {
		printLog(info, msg...)
	}
}

func D(msg ...interface{}) {
	if defaultLevel >= LevelDebug {
		printLog(debug, msg...)
	}
}

var logMutex sync.Mutex

func printLog(tag logTag, msg ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	fmt.Println("["+defaultLogName+"-"+string(tag)+"]", date.Format(time.Now(), timeFormat), fmt.Sprint(msg...))
}
