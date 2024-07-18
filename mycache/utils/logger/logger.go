package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	if Logger != nil {
		return
	}
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// dir, err := os.Getwd()
	// if err != nil {
	// 	log.Printf("get directory failed:%v", err)
	// 	return
	// }
	// logFilePath := dir + "/logs/"

	// _, err = os.Stat(logFilePath)
	// if os.IsNotExist(err) {
	// 	if err := os.MkdirAll(logFilePath, 0777); err != nil {
	// 		log.Println(err.Error())
	// 		return
	// 	}
	// }

	// fileName := filepath.Join(logFilePath, time.Now().Format("2006-01-02")+".log")
	// out, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	// if err != nil {
	// 	log.Printf("open log file failed:%v", err)
	// 	return
	// }
	logger.Out = os.Stdout

	Logger = logger
}
