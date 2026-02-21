package utils

import (
	"log-handler/internal/parser"
	"log-handler/internal/scanner"
	"log/slog"
	"sync"
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

func ProcessFilesConcurrently(filePaths []string, numWorkers int) ([]parser.LogEntry, error) {
	jobs := make(chan string, len(filePaths))
	results := make(chan []parser.LogEntry)
	wg := &sync.WaitGroup{}

	for _, v := range filePaths {
		jobs <- v
	}
	close(jobs)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go fileWorker(jobs, results, wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var res []parser.LogEntry

	for log := range results {
		res = append(res, log...)
	}

	return res, nil
}

func fileWorker(jobs <-chan string, results chan<- []parser.LogEntry, wg *sync.WaitGroup) {
	defer wg.Done()

	for pth := range jobs {
		logs, err := scanner.ReadLogFile(pth)
		if err != nil {
			slog.Error("failed to read log file", "error", err, "path", pth)
			continue
		}
		results <- logs
	}
}
