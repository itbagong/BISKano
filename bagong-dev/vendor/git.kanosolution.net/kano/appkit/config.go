package appkit

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/logger"
	"github.com/spf13/viper"
)

func ReadConfig(configPath string, dest interface{}) error {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("unable to read config file. %s", err.Error())
	}

	if err := v.Unmarshal(dest); err != nil {
		return fmt.Errorf("unable to parse config file. %s", err.Error())
	}

	return nil
}

func Log() *logger.LogEngine {
	l := logger.NewLogEngine(true, false, "", "", "")
	l.SetStdoutTemplate(func(li logger.LogItem) string {
		_, file, line := getCaller(7) // default 6 for codekit.LogEngine
		sourceFile := fmt.Sprintf("(%v#%v)", file, line)
		logTypeFmt := string(li.LogType)
		if li.LogType == "INFO" {
			logTypeFmt = aurora.BrightBlue(li.LogType).String()
		} else if li.LogType == "ERROR" {
			logTypeFmt = aurora.BgRed(aurora.Yellow(li.LogType)).String()
		} else if li.LogType == "WARNING" {
			logTypeFmt = aurora.Yellow(li.LogType).String()
		} else if li.LogType == "DEBUG" {
			logTypeFmt = aurora.Gray(10, li.LogType).String()
		}
		linebreak := strings.HasSuffix(li.Msg, "\n")
		li.Msg = strings.TrimSuffix(li.Msg, "\n")
		m := fmt.Sprintf("%v %s %s %s",
			aurora.BrightCyan(codekit.Date2String(time.Now(), "YYYY-MM-dd HH:mm:ss TS")),
			aurora.BrightBlue(logTypeFmt),
			li.Msg,
			aurora.Gray(10, sourceFile),
		)
		if linebreak {
			m += "\n"
		}
		return m
	})
	return l
}

func LogWithPrefix(prefix string) *logger.LogEngine {
	l := logger.NewLogEngine(true, false, "", "", "")

	l.SetStdoutTemplate(func(li logger.LogItem) string {
		logTypeFmt := string(li.LogType)
		if li.LogType == "INFO" {
			logTypeFmt = aurora.BrightBlue(li.LogType).String()
		} else if li.LogType == "ERROR" {
			logTypeFmt = aurora.BgRed(aurora.Yellow(li.LogType)).String()
		} else if li.LogType == "WARNING" {
			logTypeFmt = aurora.Yellow(li.LogType).String()
		}

		return fmt.Sprintf("%v %s %s %s",
			aurora.BrightCyan(codekit.Date2String(time.Now(), "YYYY-MM-dd HH:mm:ss TS")),
			aurora.BrightBlue(logTypeFmt),
			aurora.Green(prefix),
			li.Msg)
	})

	return l
}

func MakeID(prefix string, l int) string {
	if l < 20 {
		l = 20
	}
	// p1 := codekit.Date2String(time.Now(), "YYMMddHHmmss")
	// 2006-01-02T15:04:05.000000000Z07:00 // support nano, time 23 char, 24++ random string
	p1 := strings.Replace(time.Now().Format("20060102150405.000000000"), ".", "", 1)
	lp2 := l - len(prefix) - len(p1)
	if lp2 <= 0 {
		if prefix == "" {
			return p1
		}
		return fmt.Sprintf("%s%s", prefix, p1)
	}
	p2 := codekit.GenerateRandomString("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", lp2)
	if prefix == "" {
		return fmt.Sprintf("%s%s", p1, p2)
	}
	return fmt.Sprintf("%s%s%s", prefix, p1, p2)
}

func getCaller(skip int) (uintptr, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return 0, "", 0
	}

	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}

	return pc, file, line
}
