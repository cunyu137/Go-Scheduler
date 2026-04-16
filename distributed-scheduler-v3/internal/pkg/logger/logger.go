package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New(component string) *logrus.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	l.SetLevel(logrus.InfoLevel)
	return l
}
