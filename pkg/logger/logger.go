package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	SugaredLogger *zap.SugaredLogger
	Env      string
	DingUrl  string
)

type LogConfig struct {
	LogLevel string
	LogFile  string
	IsDebug  int //0-not debug(output only file)  1-debug (output either file and stdout )
}

func InitLogger(conf *LogConfig,env,dingUrl string) error {
	isDebug := true
	if conf.IsDebug != 1 {
		isDebug = false
	}
	err := initLogger(conf.LogFile, conf.LogLevel, isDebug)
	if err != nil {
		return err
	}
	log.SetFlags(log.Lmicroseconds | log.Lshortfile | log.LstdFlags)
	Env,DingUrl = env,dingUrl
	return nil
}

func ZnTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func initLogger(path string, level string, isDebug bool) error {
	var js string
	if isDebug {
		js = fmt.Sprintf(`{
      "level": "%s",
      "encoding": "console",
      "outputPaths": ["stdout", "%s"],
      "errorOutputPaths": ["stdout", "%s"]
      }`, level, path, path)
	} else {
		js = fmt.Sprintf(`{
      "level": "%s",
      "encoding": "console",
      "outputPaths": ["%s"],
      "errorOutputPaths": ["%s"]
      }`, level, path, path)
	}

	var cfg zap.Config
	if err := json.Unmarshal([]byte(js), &cfg); err != nil {
		return err
	}
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeTime = ZnTimeEncoder
	var err error
	var tlog *zap.Logger
	tlog, err = cfg.Build()
	if err != nil {
		log.Fatal("init logger error: ", err)
		return err
	}
	SugaredLogger = tlog.Sugar()
	return nil
}

// Debug fmt.Sprintf to log a templated message.
func Debug(args ...interface{}) {
	SugaredLogger.Debug(args...)
}

// Info uses fmt.Sprintf to log a templated message.
func Info(args ...interface{}) {
	SugaredLogger.Info(args...)
}

// Warn uses fmt.Sprintf to log a templated message.
func Warn(args ...interface{}) {
	SendMonitor2DingDing(fmt.Sprintf("%v",args))
	SugaredLogger.Warn(args...)
}

// Error uses fmt.Sprintf to log a templated message.
func Error(args ...interface{}) {
	SendMonitor2DingDing(fmt.Sprintf("%v",args))
	SugaredLogger.Error(args...)
}

// Debugf fmt.Sprintf to log a templated message.
func Debugf(format string, args ...interface{}) {
	SugaredLogger.Debugf(format, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(format string, args ...interface{}) {
	SugaredLogger.Infof(format, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(format string, args ...interface{}) {
	SendMonitor2DingDing(fmt.Sprintf(format, args))
	SugaredLogger.Warnf(format, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(format string, args ...interface{}) {
	SendMonitor2DingDing(fmt.Sprintf(format, args))
	SugaredLogger.Errorf(format, args...)
}

func SendMonitor2DingDing(args string) {
	//生产环境才会发送钉钉信息
	if Env != "pro" {
		return
	}

	if len(DingUrl) != 0 {
		http.Post(DingUrl, "application/json", strings.NewReader(args))
	}
}