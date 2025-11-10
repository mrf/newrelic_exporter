package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		level  logrus.Level
	}{
		{
			name:   "Debug level",
			config: Config{Level: DebugLevel},
			level:  logrus.DebugLevel,
		},
		{
			name:   "Info level",
			config: Config{Level: InfoLevel},
			level:  logrus.InfoLevel,
		},
		{
			name:   "Warn level",
			config: Config{Level: WarnLevel},
			level:  logrus.WarnLevel,
		},
		{
			name:   "Error level",
			config: Config{Level: ErrorLevel},
			level:  logrus.ErrorLevel,
		},
		{
			name:   "Default level",
			config: Config{Level: LogLevel("invalid")},
			level:  logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup(tt.config)
			if log.GetLevel() != tt.level {
				t.Errorf("Expected level %v, got %v", tt.level, log.GetLevel())
			}
		})
	}
}

func TestJSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:      InfoLevel,
		JSONFormat: true,
		Output:     buf,
	})

	Info("test message")

	output := buf.String()
	if !strings.Contains(output, `"level":"info"`) {
		t.Error("Expected JSON format with level field")
	}
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Error("Expected JSON format with msg field")
	}
}

func TestTextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:      InfoLevel,
		JSONFormat: false,
		Output:     buf,
	})

	Info("test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Expected text format with INFO level")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Expected text format with message")
	}
}

func TestDebugLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  DebugLevel,
		Output: buf,
	})

	Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message not logged")
	}
}

func TestDebugfLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  DebugLevel,
		Output: buf,
	})

	Debugf("debug %s", "formatted")

	output := buf.String()
	if !strings.Contains(output, "debug formatted") {
		t.Error("Formatted debug message not logged")
	}
}

func TestInfoLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  InfoLevel,
		Output: buf,
	})

	Info("info message")

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Error("Info message not logged")
	}
}

func TestInfofLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  InfoLevel,
		Output: buf,
	})

	Infof("info %s", "formatted")

	output := buf.String()
	if !strings.Contains(output, "info formatted") {
		t.Error("Formatted info message not logged")
	}
}

func TestWarnLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  WarnLevel,
		Output: buf,
	})

	Warn("warning message")

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Error("Warning message not logged")
	}
}

func TestWarnfLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  WarnLevel,
		Output: buf,
	})

	Warnf("warning %s", "formatted")

	output := buf.String()
	if !strings.Contains(output, "warning formatted") {
		t.Error("Formatted warning message not logged")
	}
}

func TestErrorLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  ErrorLevel,
		Output: buf,
	})

	Error("error message")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Error("Error message not logged")
	}
}

func TestErrorfLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  ErrorLevel,
		Output: buf,
	})

	Errorf("error %s", "formatted")

	output := buf.String()
	if !strings.Contains(output, "error formatted") {
		t.Error("Formatted error message not logged")
	}
}

func TestWithField(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:      InfoLevel,
		JSONFormat: true,
		Output:     buf,
	})

	WithField("key", "value").Info("message")

	output := buf.String()
	if !strings.Contains(output, `"key":"value"`) {
		t.Error("Field not included in log output")
	}
}

func TestWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:      InfoLevel,
		JSONFormat: true,
		Output:     buf,
	})

	WithFields(logrus.Fields{
		"key1": "value1",
		"key2": "value2",
	}).Info("message")

	output := buf.String()
	if !strings.Contains(output, `"key1":"value1"`) {
		t.Error("Field key1 not included in log output")
	}
	if !strings.Contains(output, `"key2":"value2"`) {
		t.Error("Field key2 not included in log output")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  InfoLevel,
		Output: buf,
	})

	Debug("debug message")
	Info("info message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not be logged at Info level")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message should be logged at Info level")
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger returned nil")
	}
	if logger != log {
		t.Error("GetLogger did not return the package logger")
	}
}

func TestPrintFunctions(t *testing.T) {
	buf := &bytes.Buffer{}
	Setup(Config{
		Level:  InfoLevel,
		Output: buf,
	})

	Print("print message")
	output := buf.String()
	if !strings.Contains(output, "print message") {
		t.Error("Print message not logged")
	}

	buf.Reset()
	Printf("print %s", "formatted")
	output = buf.String()
	if !strings.Contains(output, "print formatted") {
		t.Error("Printf message not logged")
	}
}
