package logentry

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"time"
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
		return LogEntry{}, fmt.Errorf("request_id not found in log line")
	}

	usrId := userRegex.FindStringSubmatch(res[4])
	if len(usrId) != 2 {
		return LogEntry{}, fmt.Errorf("user_id not found in log line")
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
	result := make([]LogEntry, 0, 100)
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
