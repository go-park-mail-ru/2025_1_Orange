package logger

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type CoolFormatter struct {
	TimestampFormat string
	HideKeys        bool
	NoColors        bool
	TrimMessages    bool
}

func (f *CoolFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := getLevelColor(entry.Level)

	timeStampFormat := f.TimestampFormat
	if timeStampFormat == "" {
		timeStampFormat = time.DateTime
	}
	b := &bytes.Buffer{}

	b.WriteString(entry.Time.Format(timeStampFormat))

	level := strings.ToUpper(entry.Level.String())

	if !f.NoColors {
		fmt.Fprintf(b, "\x1b[%dm[%s] ", levelColor, level)
	} else {
		fmt.Fprintf(b, "[%s] ", level)
	}

	// не уверен, стоит ли так делать
	// но это гарантирует порядок следеования id
	if requestID, ok := entry.Data["requestID"].(string); ok {
<<<<<<< HEAD
<<<<<<< HEAD
>>>>>>> 2e508df (Added logger.)
=======
>>>>>>> 2100c7a (Add migrations for vacancies)
=======
>>>>>>> 2100c7a561956d1c7aba82955674aaffa4c2399e
		b.WriteString("[RID=")
		b.WriteString(requestID)
		b.WriteString("] ")
		delete(entry.Data, "requestID")
	}

	packageName, funcName := getCallerInfo()
	b.WriteString("[")
	b.WriteString(packageName)
	b.WriteString(".")
	b.WriteString(funcName)
	b.WriteString("] ")

	if f.TrimMessages {
		b.WriteString(strings.TrimSpace(entry.Message))
	} else {
		b.WriteString(entry.Message)
	}
	b.WriteString(" ")

	f.WriteFields(b, entry, levelColor)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *CoolFormatter) WriteFields(b *bytes.Buffer, entry *logrus.Entry, color int) {
	for key, value := range entry.Data {
		if f.HideKeys {
			fmt.Fprintf(b, "%v ", value)
		} else {
			if f.NoColors {
				fmt.Fprintf(b, "%s=%v ", key, value)
			} else {
				fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m=%v ", color, key, value)
			}
		}
	}
}

const (
	BLACK   = 30
	RED     = 31
	GREEN   = 32
	YELLOW  = 33
	BLUE    = 34
	MAGENTA = 35
	CYAN    = 36
	WHITE   = 37
	DEFAULT = 0
)

func getLevelColor(level logrus.Level) int {
	switch level {
	case logrus.InfoLevel:
		return BLUE
	case logrus.DebugLevel, logrus.TraceLevel:
		return WHITE
	case logrus.WarnLevel:
		return YELLOW
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return RED
	default:
		return DEFAULT
	}
}

// тут тоже сомнительно
// skip всегда разный и надо пробегать циклом по стеку вызовов
func getCallerInfo() (string, string) {
	for skip := 3; skip < 10; skip++ {
		pc, _, _, ok := runtime.Caller(skip)
		if !ok {
			break
		}

		fullFuncName := runtime.FuncForPC(pc).Name()

		if !strings.Contains(fullFuncName, "logrus") {
			parts := strings.Split(fullFuncName, "/")
			lastPart := parts[len(parts)-1]

			pkgFunc := strings.Split(lastPart, ".")
			if len(pkgFunc) >= 2 {
				return pkgFunc[0], pkgFunc[1]
			}
			return "unknown", fullFuncName
		}
	}
	return "unknown", "unknown"
}

var Log = logrus.New()

func init() {
	Log.SetFormatter(&CoolFormatter{
		TrimMessages: true,
	})
}
