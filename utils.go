package main

import (
	"runtime"

	"github.com/sirupsen/logrus"
)

func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func logFatal(err error, fct string, msg string, param string) {
	logrus.WithFields(logrus.Fields{
		"param": param,
		"error": err,
		"fct":   fct,
	}).Warn(msg)
	log.WithFields(logrus.Fields{
		"param": param,
		"error": err,
		"fct":   fct,
	}).Fatal(msg)
}

func logDebug(fct string, msg string, param string) {
	log.WithFields(logrus.Fields{
		"param": param,
		"fct":   fct,
	}).Debug(msg)
}

func logWarn(fct string, msg string, param string) {
	logrus.WithFields(logrus.Fields{
		"param": param,
	}).Warn(msg)
	log.WithFields(logrus.Fields{
		"param": param,
		"fct":   fct,
	}).Warn(msg)
}

func logInfo(fct string, msg string, param string) {
	logrus.WithFields(logrus.Fields{
		"param": param,
	}).Info(msg)
	log.WithFields(logrus.Fields{
		"param": param,
		"fct":   fct,
	}).Info(msg)
}

func checkErr(err error, fct string, msg string, param string) {
	if err != nil {
		logFatal(err, msg, fct, param)
	}
}
