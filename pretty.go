package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	initialIndent   = 4
	indentIncrement = 2
)

type prettyOptions struct {
	Colorize   bool
	LevelFirst bool
	TimeKey    string
	MessageKey string
	LevelKey   string
}

func pretty(input []byte, options *prettyOptions) string {
	if options == nil {
		options = &prettyOptions{}
	}

	timeKey := "time"
	levelKey := "level"
	msgKey := "msg"

	if options.TimeKey != "" {
		timeKey = options.TimeKey
	}
	if options.LevelKey != "" {
		levelKey = options.LevelKey
	}
	if options.MessageKey != "" {
		msgKey = options.MessageKey
	}

	var logData map[string]interface{}
	if err := json.Unmarshal(input, &logData); err != nil {
		return fmt.Sprintf("%s\n", input)
	}

	timestamp := ""
	timeStr, _ := logData[timeKey].(string)
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		t := "[" + parsedTime.Format("15:04:05.000") + "]"
		if options.LevelFirst {
			timestamp = " " + t
		} else {
			timestamp = t + " "
		}
	}

	level, _ := logData[levelKey].(string)
	msg, _ := logData[msgKey].(string)

	if options.Colorize {
		level, msg = colorize(level, msg)
	}

	var builder strings.Builder
	for key, value := range logData {
		if key == timeKey || key == levelKey || key == msgKey {
			continue
		}

		currentIndent := initialIndent
		indentation := strings.Repeat(" ", currentIndent)
		if m, ok := value.(map[string]interface{}); ok {
			builder.WriteString(fmt.Sprintf("%s%s: {\n", indentation, key))
			processMap(&builder, m, currentIndent+indentIncrement)
			builder.WriteString(fmt.Sprintf("%s}\n", indentation))
		} else {
			builder.WriteString(fmt.Sprintf("%s%s: \"%s\"\n", indentation, key, value))
		}
	}

	if options.LevelFirst {
		return fmt.Sprintf("%s%s: %s\n%s", level, timestamp, msg, builder.String())
	}
	return fmt.Sprintf("%s%s: %s\n%s", timestamp, level, msg, builder.String())
}

func colorize(lvl, msg string) (level, message string) {
	level = fmt.Sprintf("\u001B[32m%s\u001B[39m", lvl)
	message = fmt.Sprintf("\u001B[36m%s\u001B[39m", msg)
	return
}

func processMap(b *strings.Builder, m map[string]interface{}, currentIndent int) {
	indentation := strings.Repeat(" ", currentIndent)
	for key, value := range m {
		if m, ok := value.(map[string]interface{}); ok {
			b.WriteString(fmt.Sprintf("%s\"%s\": {\n", indentation, key))
			processMap(b, m, currentIndent+indentIncrement)
			b.WriteString(fmt.Sprintf("%s}\n", indentation))
		} else {
			b.WriteString(fmt.Sprintf("%s\"%s\": \"%s\"\n", indentation, key, value))
		}
	}
}
