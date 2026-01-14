package logger

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	loggerOnce  sync.Once
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
)

// InitLogger 初始化 (預設 Stdout)
func InitLogger() {
	// 這裡使用 Stdout 作為預設值
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetOutput 設定輸出目標
func SetOutput(w io.Writer) {
	// 1. 先確保已經初始化 (避免 nil pointer)
	loggerOnce.Do(func() {
		InitLogger()
	})

	// 2. 強制更改輸出目標為檔案
	// 重要：這會修改現有的 logger 實例，而不是建立新的
	if infoLogger != nil {
		infoLogger.SetOutput(w)
	}
	if debugLogger != nil {
		debugLogger.SetOutput(w)
	}
	if errorLogger != nil {
		errorLogger.SetOutput(w)
	}
}

// LogInfo ...
func LogInfo(v ...interface{}) {
	loggerOnce.Do(func() { InitLogger() })
	infoLogger.Println(v...)
}

// LogError ...
func LogError(v ...interface{}) {
	loggerOnce.Do(func() { InitLogger() })
	errorLogger.Println(v...)
}

// LogDebug ...
func LogDebug(v ...interface{}) {
	loggerOnce.Do(func() { InitLogger() })
	debugLogger.Println(v...)
}
