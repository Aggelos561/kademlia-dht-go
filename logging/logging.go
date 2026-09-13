package logging

import (
	"fmt"
	"log/slog"
	"os"
)

// var Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
// 	Level: slog.LevelDebug,
// }))

func EnableStdErr() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
}

func Debug(format string, v ...any) {
	// slog.SetDefault(Logger)
	slog.Debug(fmt.Sprintf(format, v...))
}

func Info(format string, v ...any) {
	// slog.SetDefault(Logger)
	slog.Info(fmt.Sprintf(format, v...))
}

func Warning(format string, v ...any) {
	// slog.SetDefault(Logger)
	slog.Warn(fmt.Sprintf(format, v...))
}

func Error(format string, v ...any) {
	// slog.SetDefault(Logger)
	slog.Error(fmt.Sprintf(format, v...))
}
