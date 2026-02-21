package main

import (
	"fmt"
	"log-handler/internal/cli"
	"log-handler/internal/processor"
	"log-handler/internal/reporter"
	"log-handler/internal/scanner"
	"log-handler/internal/utils"
	"log/slog"
	"time"
)

func main() {
	startTime := time.Now()
	dir, file, err := cli.ParseCommandLineArgs()
	if err != nil {
		slog.Error("parse commands error", "error", err)
		return
	}

	paths, err := scanner.ScanLogDirectory(dir)
	if err != nil {
		slog.Error("failed to scan directory", "error", err)
	}

	logs, err := utils.ProcessMultipleFiles(paths)
	if err != nil {
		slog.Error("processing files error", "error", err)
		return
	}

	correlatedReqs := processor.CorrelateRequests(logs)
	failedReqsIds := processor.DetectFailedRequests(correlatedReqs)
	failedReqsReports := make([]reporter.FailedRequestReport, len(failedReqsIds))

	for i, reqID := range failedReqsIds {

		correlReqs := correlatedReqs[reqID]
		fstFail, ok := processor.FindFirstFailure(correlReqs)
		if !ok {
			slog.Error("detect first failure error", "request_id", reqID)
			continue
		}

		failedReqIdTimeline := processor.SortTimelineByTimestamp(correlReqs)
		timeLine := make([]string, len(failedReqIdTimeline))

		for num, line := range failedReqIdTimeline {
			timeLine[num] = fmt.Sprintf("%s [%s] %s: %s", line.Timestamp, line.Level, line.Service, line.Message)
		}

		failedReqsReports[i] = reporter.FailedRequestReport{
			RequestID:      reqID,
			FailingService: fstFail.Service,
			ErrorMessage:   fstFail.Message,
			Timeline:       timeLine,
		}
	}

	finishTime := time.Since(startTime).Seconds()

	res := reporter.AnalysisResult{
		TotalEntriesProcessed: len(logs),
		FailedRequestsFound:   len(failedReqsIds),
		ProcessingTimeSeconds: finishTime,
		FailedRequests:        failedReqsReports,
	}

	if err = reporter.WriteJSONReport(res, file); err != nil {
		slog.Error("write json to file error", "error", err)
		return
	}
}
