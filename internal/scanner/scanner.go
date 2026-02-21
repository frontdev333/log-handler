package scanner

import (
	"bufio"
	"io/fs"
	"log-handler/internal/parser"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const DefaultLogCapacity = 100

func ReadLogFile(filePath string) ([]parser.LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("open file error", "error", err.Error())
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := make([]parser.LogEntry, 0, DefaultLogCapacity)
	for scanner.Scan() {
		logLine, err := parser.ParseLogLine(scanner.Text())
		if err != nil {
			slog.Error("read log line error", "error", err.Error())
			continue
		}
		result = append(result, logLine)
	}

	if scanner.Err() != nil {
		slog.Error("scanner error", "error", scanner.Err())
		return nil, scanner.Err()
	}

	return result, nil
}

func ScanLogDirectory(dirPath string) ([]string, error) {
	res := make([]string, 0, 10)
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".log") {
			res = append(res, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
