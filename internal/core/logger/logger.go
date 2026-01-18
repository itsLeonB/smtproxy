package logger

import "github.com/itsLeonB/ezutil/v2"

var Global ezutil.Logger

func Init(appNamespace string) {
	Global = ezutil.NewSimpleLogger(appNamespace, true, 0)
}

func Debug(args ...any) {
	if Global != nil {
		Global.Debug(args...)
	}
}

func Info(args ...any) {
	if Global != nil {
		Global.Info(args...)
	}
}

func Warn(args ...any) {
	if Global != nil {
		Global.Warn(args...)
	}
}

func Error(args ...any) {
	if Global != nil {
		Global.Error(args...)
	}
}

func Fatal(args ...any) {
	if Global != nil {
		Global.Fatal(args...)
	}
}

func Debugf(format string, args ...any) {
	if Global != nil {
		Global.Debugf(format, args...)
	}
}

func Infof(format string, args ...any) {
	if Global != nil {
		Global.Infof(format, args...)
	}
}

func Warnf(format string, args ...any) {
	if Global != nil {
		Global.Warnf(format, args...)
	}
}

func Errorf(format string, args ...any) {
	if Global != nil {
		Global.Errorf(format, args...)
	}
}

func Fatalf(format string, args ...any) {
	if Global != nil {
		Global.Fatalf(format, args...)
	}
}
