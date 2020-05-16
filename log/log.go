package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logger = logrus.New()

func init()  {
	Logger.SetFormatter(&logrus.TextFormatter{})
	file, _ := os.OpenFile("logs/homerover.log", os.O_CREATE|os.O_WRONLY, 0666)
	Logger.SetOutput(file)
	Logger.SetLevel(logrus.InfoLevel)
}
