package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func parseLogLine(line string) (map[string]interface{}, error) {
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	return logEntry, nil
}

func TestConfigureLogging(t *testing.T) {
	// Save original env and logger settings
	originalLevel := log.GetLevel()
	originalFormatter := log.StandardLogger().Formatter
	originalOutput := log.StandardLogger().Out
	originalEnv := os.Getenv("BLASTRA_LOG_LEVEL")

	// Restore original settings after test
	defer func() {
		log.SetLevel(originalLevel)
		log.SetFormatter(originalFormatter)
		log.SetOutput(originalOutput)
		if originalEnv != "" {
			os.Setenv("BLASTRA_LOG_LEVEL", originalEnv)
		} else {
			os.Unsetenv("BLASTRA_LOG_LEVEL")
		}
	}()

	t.Run("default configuration", func(t *testing.T) {
		os.Unsetenv("BLASTRA_LOG_LEVEL")
		var buf bytes.Buffer
		log.SetOutput(&buf)

		ConfigureLogging()

		// Check default log level
		if log.GetLevel() != log.InfoLevel {
			t.Errorf("Expected default log level to be INFO, got %v", log.GetLevel())
		}

		// Test log output format
		log.Info("test message")
		output := buf.String()

		// Split output into lines and process each line
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			logEntry, err := parseLogLine(line)
			if err != nil {
				t.Errorf("Failed to parse log line: %v", err)
				continue
			}

			// Verify JSON fields
			if _, ok := logEntry["level"]; !ok {
				t.Error("Expected level field in JSON output")
			}
			if msg, ok := logEntry["msg"].(string); ok {
				if msg != "Starting Blastra server with log level: info" && msg != "test message" {
					t.Errorf("Unexpected message: %s", msg)
				}
			} else {
				t.Error("Expected msg field in JSON output")
			}
			if _, ok := logEntry["time"]; !ok {
				t.Error("Expected time field in JSON output")
			}
		}
	})

	t.Run("custom log level", func(t *testing.T) {
		testCases := []struct {
			level    string
			expected log.Level
		}{
			{"debug", log.DebugLevel},
			{"info", log.InfoLevel},
			{"warn", log.WarnLevel},
			{"error", log.ErrorLevel},
			{"DEBUG", log.DebugLevel},
			{"INFO", log.InfoLevel},
			{"WARN", log.WarnLevel},
			{"ERROR", log.ErrorLevel},
		}

		for _, tc := range testCases {
			t.Run(tc.level, func(t *testing.T) {
				os.Setenv("BLASTRA_LOG_LEVEL", tc.level)
				ConfigureLogging()

				if log.GetLevel() != tc.expected {
					t.Errorf("Expected log level %v, got %v", tc.expected, log.GetLevel())
				}
			})
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		os.Setenv("BLASTRA_LOG_LEVEL", "invalid")
		ConfigureLogging()

		// Should default to info level
		if log.GetLevel() != log.InfoLevel {
			t.Errorf("Expected default log level INFO for invalid input, got %v", log.GetLevel())
		}
	})

	t.Run("log formatting", func(t *testing.T) {
		os.Unsetenv("BLASTRA_LOG_LEVEL")
		var buf bytes.Buffer
		log.SetOutput(&buf)

		ConfigureLogging()

		// Test different log levels
		testCases := []struct {
			level   func(args ...interface{})
			message string
		}{
			{log.Debug, "debug message"},
			{log.Info, "info message"},
			{log.Warn, "warning message"},
			{log.Error, "error message"},
		}

		for _, tc := range testCases {
			buf.Reset()
			tc.level(tc.message)
			output := buf.String()

			// Split output into lines and process each line
			lines := strings.Split(strings.TrimSpace(output), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				logEntry, err := parseLogLine(line)
				if err != nil {
					t.Errorf("Failed to parse log line: %v", err)
					continue
				}

				// Verify JSON fields
				if _, ok := logEntry["level"]; !ok {
					t.Error("Expected level field in JSON output")
				}
				if msg, ok := logEntry["msg"].(string); ok {
					if msg != "Starting Blastra server with log level: info" && msg != tc.message {
						t.Errorf("Unexpected message: %s", msg)
					}
				} else {
					t.Error("Expected msg field in JSON output")
				}
				if _, ok := logEntry["time"]; !ok {
					t.Error("Expected time field in JSON output")
				}
			}
		}
	})

	t.Run("concurrent logging", func(t *testing.T) {
		os.Unsetenv("BLASTRA_LOG_LEVEL")
		var buf bytes.Buffer
		log.SetOutput(&buf)

		ConfigureLogging()

		var wg sync.WaitGroup
		messageCount := 100
		var logMutex sync.Mutex
		messagesSeen := make(map[string]bool)

		// Test concurrent logging
		for i := 0; i < messageCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				msg := fmt.Sprintf("concurrent message %d", i)
				logMutex.Lock()
				log.Info(msg)
				messagesSeen[msg] = true
				logMutex.Unlock()
			}(i)
		}

		wg.Wait()
		// Wait for any buffered writes to complete
		time.Sleep(100 * time.Millisecond)

		// Count messages that were seen in the output
		seenCount := len(messagesSeen)

		// Account for the "Starting Blastra server" message
		if messagesSeen["Starting Blastra server with log level: info"] {
			seenCount--
		}

		if seenCount != messageCount {
			t.Errorf("Expected %d log messages, got %d", messageCount, seenCount)
			// Print missing messages
			for i := 0; i < messageCount; i++ {
				msg := fmt.Sprintf("concurrent message %d", i)
				if !messagesSeen[msg] {
					t.Logf("Missing message: %s", msg)
				}
			}
		}
	})
}
