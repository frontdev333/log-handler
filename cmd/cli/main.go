package main

import (
	"fmt"
	"log-handler/internal/commandLine"
	"log-handler/internal/jsonHandler"
	"log-handler/internal/logentry"
	"log/slog"
	"time"
)

func main() {
	startTime := time.Now()
	dir, file, err := commandLine.ParseCommandLineArgs()
	if err != nil {
		slog.Error("parse commands error", "error", err)
		return
	}
	paths, err := logentry.ScanLogDirectory(dir)
	if err != nil {
		slog.Error("failed to scan directory", "error", err)
	}
	logs, err := logentry.ProcessMultipleFiles(paths)
	if err != nil {
		slog.Error("processing files error", "error", err)
		return
	}

	correlatedReqs := logentry.CorrelateRequests(logs)

	failedReqsIds := logentry.DetectFailedRequests(correlatedReqs)

	failedReqsReports := make([]jsonHandler.FailedRequestReport, len(failedReqsIds))

	for i, reqID := range failedReqsIds {

		correlReqs := correlatedReqs[reqID]
		fstFail, ok := logentry.FindFirstFailure(correlReqs)
		if !ok {
			slog.Error("detect first failure error", "request_id", reqID)
			continue
		}

		failedReqIdTimeline := logentry.SortTimelineByTimestamp(correlReqs)
		timeLine := make([]string, len(failedReqIdTimeline))

		for num, line := range failedReqIdTimeline {
			timeLine[num] = fmt.Sprintf("%s [%s] %s: %s", line.Timestamp, line.Level, line.Service, line.Message)
		}

		failedReqsReports[i] = jsonHandler.FailedRequestReport{
			RequestID:      reqID,
			FailingService: fstFail.Service,
			ErrorMessage:   fstFail.Message,
			Timeline:       timeLine,
		}
	}

	finishTime := time.Since(startTime).Seconds()

	res := jsonHandler.AnalysisResult{
		TotalEntriesProcessed: len(logs),
		FailedRequestsFound:   len(failedReqsIds),
		ProcessingTimeSeconds: finishTime,
		FailedRequests:        failedReqsReports,
	}

	if err = jsonHandler.WriteJSONReport(res, file); err != nil {
		slog.Error("write json to file error", "error", err)
		return
	}
}
