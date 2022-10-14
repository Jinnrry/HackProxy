package log

import (
	"fmt"
	"runtime"
	"time"
)

type LogLevel byte

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var levelNames = []string{
	LevelTrace: "Trace",
	LevelDebug: "Debug",
	LevelInfo:  "Info",
	LevelWarn:  "Warn",
	LevelError: "Error",
	LevelFatal: "Fatal",
}

type HackLog struct {
	Level LogLevel
}

var instance *HackLog

func init() {
	instance = &HackLog{Level: LevelInfo}
}

// 设置日志级别
func SetLogLevel(level LogLevel) {
	instance.Level = level
}

func GetLogLevel() LogLevel {
	return instance.Level
}

func Trace(v ...any) {
	level := LevelTrace

	filename, line := "???", 0
	_, filename, line, _ = runtime.Caller(1)
	if instance.Level <= level {
		fmt.Printf("[%s][%s][%s:%d]:%s\n", levelNames[level], time.Now().Format("2006-01-02 15:04:05"), filename, line, fmt.Sprint(v...))
	}
}

func Debug(v ...any) {
	level := LevelDebug
	filename, line := "???", 0
	_, filename, line, _ = runtime.Caller(1)
	if instance.Level <= level {
		fmt.Printf("[%s][%s][%s:%d]:%s\n", levelNames[level], time.Now().Format("2006-01-02 15:04:05"), filename, line, fmt.Sprint(v...))
	}
}

func Info(v ...any) {
	level := LevelInfo

	filename, line := "???", 0
	_, filename, line, _ = runtime.Caller(1)
	if instance.Level <= level {
		fmt.Printf("[%s][%s][%s:%d]:%s\n", levelNames[level], time.Now().Format("2006-01-02 15:04:05"), filename, line, fmt.Sprint(v...))
	}
}

func Warn(v ...any) {
	level := LevelWarn

	filename, line := "???", 0
	_, filename, line, _ = runtime.Caller(1)
	if instance.Level <= level {
		fmt.Printf("[%s][%s][%s:%d]:%s\n", levelNames[level], time.Now().Format("2006-01-02 15:04:05"), filename, line, fmt.Sprint(v...))
	}
}

func Error(v ...any) {
	level := LevelError

	filename, line := "???", 0
	_, filename, line, _ = runtime.Caller(1)
	if instance.Level <= level {
		fmt.Printf("[%s][%s][%s:%d]:%s\n", levelNames[level], time.Now().Format("2006-01-02 15:04:05"), filename, line, fmt.Sprint(v...))
	}
}

func Fatal(v ...any) {
	level := LevelFatal

	filename, line := "???", 0
	_, filename, line, _ = runtime.Caller(1)
	if instance.Level <= level {
		fmt.Printf("[%s][%s][%s:%d]:%s\n", levelNames[level], time.Now().Format("2006-01-02 15:04:05"), filename, line, fmt.Sprint(v...))
	}
}
