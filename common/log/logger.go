package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *logrus.Logger

// InitLogger 初始化 Logger，支持每日自动备份并保留近三天日志
func InitLogger(path string) {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetLevel(logrus.InfoLevel)
	logger.SetReportCaller(true) // 启用调用者信息

	// 创建日志目录
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Printf("❌ 无法创建日志目录: %v\n", err)
		return
	}

	// 日志文件轮转配置（按天切割日志，保留最近 3 天）
	infoLog := &lumberjack.Logger{
		Filename:   path + "/info.log",
		MaxSize:    10,   // 单个日志文件最大 10MB
		MaxAge:     3,    // 只保留最近 3 天的日志
		MaxBackups: 3,    // 最多保留 3 个备份
		Compress:   true, // 启用压缩
	}

	errorLog := &lumberjack.Logger{
		Filename:   path + "/error.log",
		MaxSize:    10,
		MaxAge:     3,
		MaxBackups: 3,
		Compress:   true,
	}

	// 日志输出到文件和控制台
	multiWriterInfo := io.MultiWriter(os.Stdout, infoLog)
	multiWriterError := io.MultiWriter(os.Stdout, errorLog)

	logger.SetOutput(os.Stdout)

	// Hook: INFO 级别日志写入 info.log
	logger.AddHook(&FileHook{
		Writer: multiWriterInfo,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.WarnLevel,
		},
	})

	// Hook: ERROR 级别日志写入 error.log
	logger.AddHook(&FileHook{
		Writer: multiWriterError,
		LogLevels: []logrus.Level{
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	})

	logger.Info("✅ Logger 初始化成功，日志自动轮转，每天备份，保留最近 3 天")
}

// FileHook 自定义 Hook，支持写入日志文件
type FileHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

// Fire 实现 logrus Hook，将日志写入文件
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	entry.Data["file"], entry.Data["func"] = getCaller() // 记录调用者信息
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	line, err := formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(line)
	return err
}

// Levels 返回 Hook 适用的日志级别
func (hook *FileHook) Levels() []logrus.Level {
	return hook.LogLevels
}

// 获取调用日志的代码位置
func getCaller() (string, string) {
	for i := 3; i < 15; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		funcName := runtime.FuncForPC(pc).Name()
		if !isLogrusFile(file) {
			return fmt.Sprintf("%s:%d", trimGoPath(file), line), funcName
		}
	}
	return "unknown", "unknown"
}

// 过滤 logrus 本身的调用
func isLogrusFile(file string) bool {
	return strings.Contains(file, "logrus") || strings.Contains(file, "logger.go")
}

// 去掉 GOPATH 或 module path，优化 file 显示
func trimGoPath(file string) string {
	goPath := os.Getenv("GOPATH")
	if goPath != "" {
		goPathPrefix := goPath + "/src/"
		if strings.HasPrefix(file, goPathPrefix) {
			return strings.TrimPrefix(file, goPathPrefix)
		}
	}
	return file
}

// **日志方法封装**
func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// 获取 logger 实例
func GetLogger() *logrus.Logger {
	return logger
}
