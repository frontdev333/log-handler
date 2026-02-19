package main

import (
	"fmt"
	"log-handler/internal/logentry"
	"log/slog"
)

func main() {
	basePath := "./"
	fmt.Println("Scanning directory:", basePath)
	paths, err := logentry.ScanLogDirectory(basePath)
	if err != nil {
		slog.Error("failed to scan directory", "error", err)
	}
	fmt.Println("Found", len(paths), "log files:")
	for _, v := range paths {
		fmt.Println("-", v)
	}
	fmt.Println("Processing files...")
	logs, err := logentry.ProcessMultipleFiles(paths)
	if err != nil {
		slog.Error("failed to process file", "error", err)
	}

	res := logentry.CorrelateRequests(logs)
	fmt.Printf("%#v", res["req_abc123"])
}
