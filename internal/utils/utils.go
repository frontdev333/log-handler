package utils

import (
	"log-handler/internal/parser"
	"log-handler/internal/scanner"
	"log/slog"
)

func ProcessMultipleFiles(filePaths []string) ([]parser.LogEntry, error) {
	capacity := len(filePaths) * scanner.DefaultLogCapacity

	res := make([]parser.LogEntry, 0, capacity)

	for _, pth := range filePaths {
		logs, err := scanner.ReadLogFile(pth)
		if err != nil {
			slog.Error("failed to read log file", "error", err, "path", pth)
			continue
		}
		res = append(res, logs...)
	}
	return res, nil
}
