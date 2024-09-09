package debug

import (
	"log"
	"os"
)

var debugLogger *log.Logger
var doUseDebugLogger bool

func init() {
	doUseDebugLogger = true
	if doUseDebugLogger {
		file, err := os.OpenFile("/home/junchen/cni.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		debugLogger = log.New(file, "", log.LstdFlags)
	}
}

func Logf(format string, v ...any) {
	if doUseDebugLogger {
		debugLogger.Printf(format, v...)
	}
}
