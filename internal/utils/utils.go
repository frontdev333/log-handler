package utils

import (
	"context"
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

func ProcessFilesConcurrently(ctx context.Context, filePaths []string, numWorkers int) ([]parser.LogEntry, error) {
	jobs := make(chan string, len(filePaths))
	results := make(chan []parser.LogEntry)
	wg := &sync.WaitGroup{}

loop:
	for _, v := range filePaths {
		select {
		case <-ctx.Done():
			break loop
		case jobs <- v:
		}
	}
	close(jobs)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go fileWorker(ctx, wg, jobs, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var res []parser.LogEntry

	for log := range results {
		res = append(res, log...)
	}

	return res, ctx.Err()
}

func fileWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- []parser.LogEntry) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case pth, ok := <-jobs:
			if !ok {
				return
			}
			logs, err := scanner.ReadLogFile(pth)
			if err != nil {
				slog.Error("failed to read log file", "error", err, "path", pth)
				continue
			}
			select {
			case <-ctx.Done():
				return
			case results <- logs:
			}

		}
	}
}
