package processor

import (
	"log-handler/internal/parser"
	"sort"
	"strings"
)

const (
	errLevel   = "ERROR"
	warnLevel  = "WARN"
	OrphansKey = "orphans"
)

func CorrelateRequests(entries []parser.LogEntry) map[string][]parser.LogEntry {
	res := make(map[string][]parser.LogEntry)

	for _, log := range entries {
		if log.RequestID == "" {
			res[OrphansKey] = append(res[OrphansKey], log)
			continue
		}

		res[log.RequestID] = append(res[log.RequestID], log)
	}
	return res
}

// DetectFailedRequests returns a slice of requests ids
func DetectFailedRequests(correlatedRequests map[string][]parser.LogEntry) []string {
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

func FindFirstFailure(requestEntries []parser.LogEntry) (parser.LogEntry, bool) {
	res := make([]parser.LogEntry, 0)

	for _, v := range requestEntries {
		level := v.Level
		if strings.Contains(level, errLevel) || strings.Contains(level, warnLevel) {
			res = append(res, v)
		}
	}

	res = SortTimelineByTimestamp(res)

	if len(res) > 0 {
		return res[0], true
	}

	return parser.LogEntry{}, false
}

func SortTimelineByTimestamp(entries []parser.LogEntry) []parser.LogEntry {
	tmpEntries := make([]parser.LogEntry, len(entries))

	copy(tmpEntries, entries)

	sort.Slice(tmpEntries, func(i, j int) bool {
		fstVal := tmpEntries[i].Timestamp
		scndVal := tmpEntries[j].Timestamp
		return fstVal.Before(scndVal)
	})

	return tmpEntries
}
