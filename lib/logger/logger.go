package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Settings 存储日志配置
type Settings struct {
	Path       string `yaml:"path"`        // 日志文件路径
	Name       string `yaml:"name"`        // 日志文件名
	Ext        string `yaml:"ext"`         // 日志文件扩展名
	TimeFormat string `yaml:"time-format"` // 日志文件时间格式
}

// 定义日志文件

var (
	logFile            *os.File
	defaultPrefix      = ""                                                  // 默认前缀
	defaultCallerDepth = 2                                                   // 默认调用深度
	logger             *log.Logger                                           // 日志记录器
	mu                 sync.Mutex                                            // 互斥锁
	logPrefix          = ""                                                  // 日志前缀
	levelFlags         = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"} // 日志级别
)

// 日志级别
type logLevel int

const (
	DEBUG   logLevel = iota // 0
	INFO                    // 1
	WARNING                 // 2
	ERROR                   // 3
	FATAL                   // 4
)

// 日志标志
const flags = log.LstdFlags

func init() {
	// 初始化日志记录器
	logger = log.New(os.Stdout, defaultPrefix, flags)
}

// Setup 初始化日志
func Setup(settings *Settings) {
	var err error
	dir := settings.Path
	fileName := fmt.Sprintf("%s-%s.%s",
		settings.Name,
		time.Now().Format(settings.TimeFormat),
		settings.Ext)
	logFile, err = mustOpen(fileName, dir)
	if err != nil {
		log.Fatalf("logging.Setup err: %s", err)
	}

	// 设置日志输出
	mw := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(mw, defaultPrefix, flags)

}
func setPrefix(level logLevel) {
	// 获取当前调用的文件名和行号
	_, file, line, ok := runtime.Caller(defaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d] ", levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s] ", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}

// Debug 打印调试日志
func Debug(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(DEBUG)
	logger.Println(v...)
}

// Info 打印信息日志
func Info(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(INFO)
	logger.Println(v...)
}

// Warn 打印警告日志
func Warn(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(WARNING)
	logger.Println(v...)
}

// Error 打印错误日志
func Error(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(ERROR)
	logger.Println(v...)
}

// Fatal 打印错误日志，然后停止程序
func Fatal(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	// setPrefix 是一个内部函数，用于设置日志前缀
	setPrefix(FATAL)
	logger.Fatalln(v...)
}
