package logger

import (
	"bytes"
	// "fmt"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCoolFormatter_Format(t *testing.T) {
	t.Parallel()

	now := time.Now()
	formatter := &CoolFormatter{
		TimestampFormat: time.RFC3339,
		TrimMessages:    true,
	}

	t.Run("Basic formatting", func(t *testing.T) {
		t.Parallel()

		entry := &logrus.Entry{
			Level:   logrus.InfoLevel,
			Time:    now,
			Message: "test message",
			Data:    logrus.Fields{"key": "value"},
		}

		result, err := formatter.Format(entry)
		require.NoError(t, err)
		resultStr := stripAnsi(string(result))

		require.Contains(t, resultStr, now.Format(time.RFC3339))
		require.Contains(t, resultStr, "[INFO]")
		require.Contains(t, resultStr, "test message")
		require.Contains(t, resultStr, "key=value")
	})

	t.Run("With request ID", func(t *testing.T) {
		t.Parallel()

		entry := &logrus.Entry{
			Level:   logrus.InfoLevel,
			Time:    now,
			Message: "test message",
			Data:    logrus.Fields{"requestID": "12345", "key": "value"},
		}

		result, err := formatter.Format(entry)
		require.NoError(t, err)
		resultStr := stripAnsi(string(result))

		require.Contains(t, resultStr, "[RID=12345]")
		require.NotContains(t, resultStr, "requestID=")
		require.Contains(t, resultStr, "key=value")
	})

	t.Run("No colors", func(t *testing.T) {
		t.Parallel()

		noColorFormatter := &CoolFormatter{
			NoColors:     true,
			TrimMessages: true,
		}

		entry := &logrus.Entry{
			Level:   logrus.InfoLevel,
			Time:    now,
			Message: "test message",
			Data:    logrus.Fields{"key": "value"},
		}

		result, err := noColorFormatter.Format(entry)
		require.NoError(t, err)
		resultStr := string(result)

		require.Contains(t, resultStr, "key=value")
		require.NotContains(t, resultStr, "\x1b[")
	})
}

func TestWriteFields(t *testing.T) {
	t.Parallel()

	formatter := &CoolFormatter{}

	t.Run("Normal fields", func(t *testing.T) {
		t.Parallel()

		b := &bytes.Buffer{}
		entry := &logrus.Entry{
			Data: logrus.Fields{"key1": "value1", "key2": 42},
		}

		formatter.WriteFields(b, entry, BLUE)
		result := stripAnsi(b.String())

		require.Contains(t, result, "key1=value1")
		require.Contains(t, result, "key2=42")
	})

	t.Run("Hide keys", func(t *testing.T) {
		t.Parallel()

		hideKeysFormatter := &CoolFormatter{HideKeys: true}
		b := &bytes.Buffer{}
		entry := &logrus.Entry{
			Data: logrus.Fields{"key1": "value1"},
		}

		hideKeysFormatter.WriteFields(b, entry, BLUE)
		result := stripAnsi(b.String())

		require.Contains(t, result, "value1")
		require.NotContains(t, result, "key1=")
	})
}

// stripAnsi удаляет ANSI escape sequences из строки
func stripAnsi(str string) string {
	var re = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(str, "")
}

func TestGetLevelColor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		level  logrus.Level
		expect int
	}{
		{logrus.InfoLevel, BLUE},
		{logrus.DebugLevel, WHITE},
		{logrus.TraceLevel, WHITE},
		{logrus.WarnLevel, YELLOW},
		{logrus.ErrorLevel, RED},
		{logrus.FatalLevel, RED},
		{logrus.PanicLevel, RED},
		{logrus.Level(99), DEFAULT}, // unknown level
	}

	for _, tc := range testCases {
		result := getLevelColor(tc.level)
		require.Equal(t, tc.expect, result)
	}
}

func TestGetCallerInfo(t *testing.T) {
	t.Parallel()

	// Это сложно тестировать, так как зависит от стека вызовов
	// Можно проверить хотя бы, что возвращаются непустые значения
	pkg, fn := getCallerInfo()
	require.NotEqual(t, "unknown", pkg)
	require.NotEqual(t, "unknown", fn)
}

func TestInit(t *testing.T) {
	t.Parallel()

	// Проверяем, что в init устанавливается наш форматтер
	require.IsType(t, &CoolFormatter{}, Log.Formatter)

	// Проверяем параметры форматтера
	formatter, ok := Log.Formatter.(*CoolFormatter)
	require.True(t, ok)
	require.True(t, formatter.TrimMessages)
}
