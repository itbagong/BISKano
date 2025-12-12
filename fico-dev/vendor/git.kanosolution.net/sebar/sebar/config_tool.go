package sebar

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"git.kanosolution.net/kano/kaos"
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
			aurora.BrightCyan(codekit.Date2String(time.Now(), "YYYY-MM-dd HH:mm:ss TZ")),
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
			aurora.BrightCyan(codekit.Date2String(time.Now(), "YYYY-MM-dd HH:mm:ss TZ")),
			aurora.BrightBlue(logTypeFmt),
			aurora.Green(prefix),
			li.Msg)
	})

	return l
}

type IDPrecisionEnum string

const (
	PrecisionNone   IDPrecisionEnum = ""
	PrecisionMilli  IDPrecisionEnum = ".000"
	PrecesiionMicro IDPrecisionEnum = ".000000"
	PrecisionNano   IDPrecisionEnum = ".000000000"
)

func MakeID(prefix string, precision IDPrecisionEnum, l int) string {
	if l < 20 {
		l = 20
	}

	p1 := strings.Replace(time.Now().Format(fmt.Sprintf("20060102150405%s", precision)), ".", "", 1)
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

func ConfigHasData(cfg *AppConfig, dataElements ...string) error {
	datas := cfg.Data
	dontHave := []string{}
	for _, el := range dataElements {
		if !datas.Has(el) {
			dontHave = append(dontHave, el)
		}
	}
	if len(dontHave) > 0 {
		return fmt.Errorf("missing_config: %s", strings.Join(dontHave, ","))
	}
	return nil
}

func CopyConfigDataToService(cfg *AppConfig, s *kaos.Service, dataFields ...string) error {
	err := ConfigHasData(cfg, dataFields...)
	if err != nil {
		return err
	}
	for _, f := range dataFields {
		v := cfg.Data.Get(f, "").(string)
		s.Data().Set(f, v)
	}
	return nil
}
