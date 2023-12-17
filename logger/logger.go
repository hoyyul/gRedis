package logger

import (
	"fmt"
	"gRedis/config"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
)

type LogLevel int

type LogConfig struct {
	Path     string
	Name     string
	LogLevel LogLevel
}

var (
	DEBUG   LogLevel = 0
	INFO    LogLevel = 1
	WARNING LogLevel = 2
	PANIC   LogLevel = 3
	ERROR   LogLevel = 4
)

var (
	logLevelTable = []string{"debug", "info", "warning", "panic", "error"}
	logFile       *os.File
	logConf       *LogConfig
	logMu         *sync.Mutex
	logger        *log.Logger
	prefix        string = ""
)

func Init(config *config.Config) {
	var err error
	logConf = &LogConfig{
		Path:     config.LogDir,
		Name:     "redis.log",
		LogLevel: INFO,
	}

	logMu = &sync.Mutex{}

	for i := range logLevelTable {
		if logLevelTable[i] == config.LogLevel {
			logConf.LogLevel = LogLevel(i)
			break
		}
	}

	if _, err = os.Stat(logConf.Path); err != nil {
		if err = os.Mkdir(logConf.Path, 0755); err != nil {
			log.Panic("Failed to create log dir, ", err)
		}
	}

	fileName := path.Join(logConf.Path, logConf.Name)
	logFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Panic("Failed to create log file, ", err)
	}

	writer := io.MultiWriter(logFile, os.Stdout) // bufio是包装后的io，读写推荐用bufio
	logger = log.New(writer, "", log.LstdFlags)
}

func setPrefix(level LogLevel) {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		prefix = fmt.Sprintf("[%s][%s:%d]", logLevelTable[level], file, line)
	} else {
		prefix = fmt.Sprintf("[%s]", logLevelTable[level])
	}
	logger.SetPrefix(prefix)
}

func Debug(v ...any) {
	if DEBUG < logConf.LogLevel {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	setPrefix(DEBUG)
	logger.Println(v)
}

func Info(v ...any) {
	if INFO < logConf.LogLevel {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	setPrefix(INFO)
	logger.Println(v)
}

func Warning(v ...any) {
	if WARNING < logConf.LogLevel {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	setPrefix(WARNING)
	logger.Println(v)
}

func Panic(v ...any) {
	if PANIC < logConf.LogLevel {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	setPrefix(PANIC)
	logger.Println(v)
}

func Error(v ...any) {
	if ERROR < logConf.LogLevel {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	setPrefix(ERROR)
	logger.Println(v)
}
