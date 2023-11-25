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
	LogLevelTable = []string{"debug", "info", "warning", "panic", "error"}
	LogConf       *LogConfig
	LogMu         sync.Mutex
	Logger        *log.Logger
	Prefix        string = ""
)

func Init(config *config.Config) {
	LogConf = &LogConfig{
		Path:     config.LogDir,
		Name:     "log",
		LogLevel: INFO,
	}

	for i := range LogLevelTable {
		if LogLevelTable[i] == config.LogLevel {
			LogConf.LogLevel = LogLevel(i)
			break
		}
	}

	if _, err := os.Stat(LogConf.Path); err != nil {
		if err := os.Mkdir(LogConf.Path, 0755); err != nil {
			log.Panic("Failed to create log dir")
		}
	}

	fileName := path.Join(LogConf.Path, LogConf.Name)
	logFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Panic("Failed to create log log file")
	}

	writer := io.MultiWriter(logFile, os.Stdout)
	Logger = log.New(writer, Prefix, log.LstdFlags)
}

func setPrefix(level LogLevel) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		Prefix = fmt.Sprintf("[%s][%s:%d]", LogLevelTable[level], file, line)
	} else {
		Prefix = fmt.Sprintf("[%s]", LogLevelTable[level])
	}
	Logger.SetPrefix(Prefix)
}

func Debug(v any) {
	if DEBUG < LogConf.LogLevel {
		return
	}
	LogMu.Lock()
	defer LogMu.Unlock()
	setPrefix(DEBUG)
	Logger.Println(v)
}

func Info(v any) {
	if INFO < LogConf.LogLevel {
		return
	}
	LogMu.Lock()
	defer LogMu.Unlock()
	setPrefix(INFO)
	Logger.Println(v)
}

func Warning(v any) {
	if WARNING < LogConf.LogLevel {
		return
	}
	LogMu.Lock()
	defer LogMu.Unlock()
	setPrefix(WARNING)
	Logger.Println(v)
}

func Panic(v any) {
	if PANIC < LogConf.LogLevel {
		return
	}
	LogMu.Lock()
	defer LogMu.Unlock()
	setPrefix(PANIC)
	Logger.Println(v)
}

func Error(v any) {
	if ERROR < LogConf.LogLevel {
		return
	}
	LogMu.Lock()
	defer LogMu.Unlock()
	setPrefix(ERROR)
	Logger.Println(v)
}
