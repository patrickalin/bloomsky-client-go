package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func logFatal(err error, fct string, msg string, params ...string) {
	logrus.WithFields(logrus.Fields{
		"param": fmt.Sprintf(strings.Join(params[:], ",")),
		"error": err,
		"fct":   fct,
	}).Warn(msg)
	log.WithFields(logrus.Fields{
		"param": fmt.Sprintf(strings.Join(params[:], ",")),
		"error": err,
		"fct":   fct,
	}).Fatal(msg)
}

func logDebug(fct string, msg string, params ...string) {
	log.WithFields(logrus.Fields{
		"param": fmt.Sprintf(strings.Join(params[:], ",")),
		"fct":   fct,
	}).Debug(msg)
}

func logWarn(fct string, msg string, params ...string) {
	logrus.WithFields(logrus.Fields{
		"param": fmt.Sprintf(strings.Join(params[:], ",")),
	}).Warn(msg)
	log.WithFields(logrus.Fields{
		"param": fmt.Sprintf(strings.Join(params[:], ",")),
		"fct":   fct,
	}).Warn(msg)
}

func logInfo(fct string, msg string, params ...string) {
	logrus.WithFields(logrus.Fields{
		"param": fmt.Sprintf(strings.Join(params[:], ",")),
	}).Info(msg)
	log.WithFields(logrus.Fields{
		"param": fmt.Sprintf(strings.Join(params[:], ",")),
		"fct":   fct,
	}).Info(msg)
}

func checkErr(err error, fct string, msg string, param string) {
	if err != nil {
		logFatal(err, msg, fct, param)
	}
}
