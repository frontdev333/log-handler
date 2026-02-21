package jsonHandler

import (
	"encoding/json"
	"log/slog"
	"os"
)

type FailedRequestReport struct {
	RequestID      string
	FailingService string
	ErrorMessage   string
	Timeline       []string
}
type AnalysisResult struct {
	TotalEntriesProcessed int
	FailedRequestsFound   int
	ProcessingTimeSeconds float64
	FailedRequests        []FailedRequestReport
}

func WriteJSONReport(result AnalysisResult, filename string) error {
	res, err := json.MarshalIndent(result, "", "	")
	if err != nil {
		slog.Error("analysis result marshal error", "error", err)
		return err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		slog.Error("open file error", "file", filename, "error", err)
		return err
	}
	defer file.Close()

	if _, err = file.Write(res); err != nil {
		slog.Error("write json to file error", "error", err)
		return err
	}
	return nil
}
