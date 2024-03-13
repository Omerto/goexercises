package file_logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"example.com/utils"
)

var timeFormat = "Jan _2 12:01:52.123"

type LogLevel int16

const (
	Undefined LogLevel = iota
	Fatal
	Error
	Warn
	Info
	Debug
)

func (l LogLevel) String() string {
	switch l {
	case Fatal:
		return "Fatal"
	case Error:
		return "Error"
	case Warn:
		return "Warn"
	case Info:
		return "Info"
	case Debug:
		return "Debug"

	}
	return "Undefined"
}

type FileLoggerInterface interface {
	LogMessage(severity LogLevel, message string, v ...interface{})
}

type fileLoggerPrivate struct {
	stopChannel    chan struct{}
	messageChannel chan string
	syncMutex      sync.Mutex
}

func newLoggerImpl() *fileLoggerPrivate {
	return &fileLoggerPrivate{
		stopChannel:    make(chan struct{}, 1), // idk why but it works only with buffereized channel
		messageChannel: make(chan string),
	}
}

func (fp *FileLogger) Stop() {
	select {
	case fp.implPrivate.stopChannel <- struct{}{}:
		close(fp.implPrivate.stopChannel)
	default:
	}
}

type FileLogger struct {
	logFile      *os.File
	tag          string
	maxRecords   uint64
	currentLevel LogLevel
	logToConsole bool
	implPrivate  *fileLoggerPrivate
}

func NewFileLogger(fileName string, level LogLevel, maxRecords uint64, toConsole bool) (*FileLogger, error) {
	openedFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("cant open file%s error:%s", fileName, err.Error())
	}

	loggerImpl := newLoggerImpl()
	instnce := &FileLogger{
		logFile:      openedFile,
		tag:          level.String(),
		maxRecords:   maxRecords,
		currentLevel: level,
		logToConsole: toConsole,
		implPrivate:  loggerImpl,
	}

	go instnce.writeToFile()

	return instnce, nil
}

func (fl *FileLogger) LogMessage(severity LogLevel, message string, v ...interface{}) {
	if severity >= fl.currentLevel {
		var currentTime = time.Now().Format(timeFormat)
		formattedMessage := currentTime + " " + fmt.Sprintf(message, v...)
		fl.implPrivate.messageChannel <- formattedMessage
	}
}

func (fl *FileLogger) LogFatal(message string, v ...interface{}) {
	fl.LogMessage(Fatal, message, v...)
	fl.LogMessage(Fatal, utils.GetCurrentStack())
}

func (fl *FileLogger) LogError(message string, v ...interface{}) {
	fl.LogMessage(Error, message, v...)
}

func (fl *FileLogger) LogWarning(message string, v ...interface{}) {
	fl.LogMessage(Warn, message, v...)
}

func (fl *FileLogger) LogInfo(message string, v ...interface{}) {
	fl.LogMessage(Info, message, v...)
}

func (fl *FileLogger) LogDebug(message string, v ...interface{}) {
	fl.LogMessage(Debug, message, v...)
}

func (fl *FileLogger) writeToFile() {
	for {
		select {
		case msg, ok := <-fl.implPrivate.messageChannel:
			if !ok {
				return
			}
			fl.implPrivate.syncMutex.Lock()
			_, err := fmt.Fprintln(fl.logFile, msg)
			// how to move this outside of lock?
			handleWriteError(err)
			if fl.logToConsole {
				_, err := fmt.Println(msg)
				handleWriteError(err)
			}
			fl.implPrivate.syncMutex.Unlock()
		case <-fl.implPrivate.stopChannel:
			return
		}
	}
}

// i have no idea what to do here
func handleWriteError(err error) {
	// if err != nil {
	// 	fmt.Printf("Error during %s:%s\n", err, err.Error())
	// }
}
