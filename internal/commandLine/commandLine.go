package commandLine

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func ParseCommandLineArgs() (inputDir string, outputFile string, err error) {
	flag.StringVar(&inputDir, "input-dir", ".", "directory to start search")
	flag.StringVar(&outputFile, "output-file", "results.json", "path to output file")
	flag.Usage = func() {
		fmt.Println(
			"Usage: log-processor [input-directory] [output-file]\n" +
				"input-directory: Directory containing .log files (default: current directory)\n" +
				"output-file: JSON output file (default: results.json)",
		)
	}

	flag.Parse()

	info, err := os.Stat(inputDir)
	if err != nil {
		slog.Error("failed to access input directory", "error", err)
		return "", "", err
	}

	if !info.IsDir() {
		msg := "given input directory path is not a directory"
		slog.Error(msg)
		return "", "", errors.New(msg)
	}

	return inputDir, outputFile, nil
}
