package logger

import "github.com/sirupsen/logrus"

func New(role string) *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	l.SetLevel(logrus.InfoLevel)
	l.WithField("role", role)
	return l
}
