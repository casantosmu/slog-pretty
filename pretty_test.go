package main

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
	"time"
)

type loggerOptions struct {
	MessageKey  string
	LevelKey    string
	MissingTime bool
}

func loggerNew(w io.Writer, options *loggerOptions) *slog.Logger {
	if options == nil {
		options = &loggerOptions{}
	}

	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if options.MissingTime {
					return slog.Attr{}
				}
				a.Value = slog.TimeValue(time.Date(2018, time.March, 30, 17, 35, 28, 992*1e6, time.UTC))
			} else if a.Key == slog.MessageKey && options.MessageKey != "" {
				a.Key = options.MessageKey
			} else if a.Key == slog.LevelKey && options.LevelKey != "" {
				a.Key = options.LevelKey
			}
			return a
		},
	}))
}

func TestPretty(t *testing.T) {
	t.Run("preserves output if not valid JSON", func(t *testing.T) {
		got := pretty([]byte("this is not json\nit's just regular output\n"), nil)
		want := "this is not json\nit's just regular output\n\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("formats a line without any extra options", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, nil)
		logger.Info("foo")

		got := pretty(b.Bytes(), nil)
		want := "[17:35:28.992] INFO: foo\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("will add color codes", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, nil)
		logger.Info("foo")

		got := pretty(b.Bytes(), &prettyOptions{
			Colorize: true,
		})
		want := "[17:35:28.992] \u001B[32mINFO\u001B[39m: \u001B[36mfoo\u001B[39m\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("can swap date and level position", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, nil)
		logger.Info("foo")

		got := pretty(b.Bytes(), &prettyOptions{
			LevelFirst: true,
		})
		want := "INFO [17:35:28.992]: foo\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("can use different message keys", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, &loggerOptions{
			MessageKey: "bar",
		})
		logger.Info("baz")

		got := pretty(b.Bytes(), &prettyOptions{
			MessageKey: "bar",
		})
		want := "[17:35:28.992] INFO: baz\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("can use different level keys", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, &loggerOptions{
			LevelKey: "bar",
		})
		logger.Warn("foo")

		got := pretty(b.Bytes(), &prettyOptions{
			LevelKey: "bar",
		})
		want := "[17:35:28.992] WARN: foo\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("works without time'", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, &loggerOptions{
			MissingTime: true,
		})
		logger.Info("foo")

		got := pretty(b.Bytes(), nil)
		want := "INFO: foo\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("prettifies properties", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, nil)
		logger.Info("foo", "a", "b")

		got := pretty(b.Bytes(), nil)
		want := "[17:35:28.992] INFO: foo\n    a: \"b\"\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("prettifies nested properties", func(t *testing.T) {
		var b bytes.Buffer
		logger := loggerNew(&b, nil)
		logger.Info("foo", slog.Group("a", slog.Group("b", "c", "d")))

		got := pretty(b.Bytes(), nil)
		want := "[17:35:28.992] INFO: foo\n    a: {\n      \"b\": {\n        \"c\": \"d\"\n      }\n    }\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}
