package logentry

import (
	"bufio"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	defaultLogCapacity = 100
	errLevel           = "ERROR"
	warnLevel          = "WARN"
)

type LogEntry struct {
	UserID    string
	RequestID string
	Message   string
	Service   string
	Level     string
	Timestamp time.Time
}

var mainRegex = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z) \[(\w+)\] (\S+):\s(.+)`)
var requestRegex = regexp.MustCompile(`request_id=([a-zA-Z0-9_]+)`)
var userRegex = regexp.MustCompile(`user_id=([a-zA-Z0-9_]+)`)
var msgRegex = regexp.MustCompile(`(.+?),`)

func ParseLogLine(line string) (LogEntry, error) {

	res := mainRegex.FindStringSubmatch(line)
	if res == nil || len(res) != 5 {
		return LogEntry{}, fmt.Errorf("incorrect log line format")
	}

	logTime, err := time.Parse("2006-01-02T15:04:05.000Z07:00", res[1])
	if err != nil {
		return LogEntry{}, err
	}

	reqId := requestRegex.FindStringSubmatch(res[4])
	if len(reqId) != 2 {
		slog.Error("request_id not found in log line", "line", line)
		reqId = []string{"", ""}
	}

	usrId := userRegex.FindStringSubmatch(res[4])
	if len(usrId) != 2 {
		slog.Warn("user_id not found in log line", "request_id", reqId[1])
		usrId = []string{"", ""}
	}

	msg := msgRegex.FindStringSubmatch(res[4])

	if len(msg) != 2 {
		return LogEntry{}, fmt.Errorf("message format invalid in log line")
	}

	return LogEntry{
		UserID:    usrId[1],
		RequestID: reqId[1],
		Message:   msg[1],
		Service:   res[3],
		Level:     res[2],
		Timestamp: logTime,
	}, nil
}

func ReadLogFile(filepath string) ([]LogEntry, error) {
	file, err := os.Open(filepath)
	if err != nil {
		slog.Error("open file error", "error", err.Error())
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := make([]LogEntry, 0, defaultLogCapacity)
	for scanner.Scan() {
		logLine, err := ParseLogLine(scanner.Text())
		if err != nil {
			slog.Error("read log line error", "error", err.Error())
			continue
		}
		result = append(result, logLine)
	}

	if scanner.Err() != nil {
		slog.Error("scanner error")
		return nil, err
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

func ProcessMultipleFiles(filePaths []string) ([]LogEntry, error) {
	capacity := len(filePaths) * defaultLogCapacity

	res := make([]LogEntry, 0, capacity)

	for _, pth := range filePaths {
		logs, err := ReadLogFile(pth)
		if err != nil {
			slog.Error("failed to read log file", "error", err, "path", pth)
			continue
		}
		res = append(res, logs...)
	}
	return res, nil
}

func CorrelateRequests(entries []LogEntry) map[string][]LogEntry {
	const orphansKey = "orphans"

	res := make(map[string][]LogEntry)

	for _, log := range entries {
		if log.RequestID == "" {
			res[orphansKey] = append(res[orphansKey], log)
			continue
		}

		res[log.RequestID] = append(res[log.RequestID], log)
	}
	return res
}

func DetectFailedRequests(correlatedRequests map[string][]LogEntry) []string {
	res := make([]string, 0)
	for reqId, v := range correlatedRequests {
		for _, log := range v {
			level := log.Level
			if strings.Contains(level, errLevel) || strings.Contains(level, warnLevel) {
				res = append(res, reqId)
				break
			}
		}
	}
	return res
}

func FindFirstFailure(requestEntries []LogEntry) (LogEntry, bool) {

	res := make([]LogEntry, 0)

	for _, v := range requestEntries {
		level := v.Level
		if strings.Contains(level, errLevel) || strings.Contains(level, warnLevel) {
			res = append(res, v)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		fstVal := res[i].Timestamp
		scndVal := res[j].Timestamp
		return fstVal.Before(scndVal)
	})

	if len(res) > 0 {
		return res[0], true
	}

	return LogEntry{}, false
}
