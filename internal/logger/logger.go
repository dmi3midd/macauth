package logger

import (
	"io"
	"log/slog"
	"os"
)

func Setup(logPath string) (*os.File, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)

	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return file, nil
}
