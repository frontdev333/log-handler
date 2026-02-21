package reporter

import (
	"encoding/json"
	"log/slog"
	"os"
)

type FailedRequestReport struct {
	RequestID      string   `json:"request_id"`
	FailingService string   `json:"failing_service"`
	ErrorMessage   string   `json:"error_message"`
	Timeline       []string `json:"timeline"`
}
type AnalysisResult struct {
	TotalEntriesProcessed int                   `json:"total_entries_processed"`
	FailedRequestsFound   int                   `json:"failed_requests_found"`
	ProcessingTimeSeconds float64               `json:"processing_time_seconds"`
	FailedRequests        []FailedRequestReport `json:"failed_requests"`
}

func WriteJSONReport(result AnalysisResult, filename string) error {
	res, err := json.MarshalIndent(result, "", "	")
	if err != nil {
		slog.Error("analysis result marshal error", "error", err)
		return err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
