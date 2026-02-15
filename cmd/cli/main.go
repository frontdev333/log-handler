package main

import (
	"fmt"
	"log-handler/internal/logentry"
)

func main() {
	log, err := logentry.ParseLogLine("2023-12-25T14:30:15.123Z [INFO] user-service: User authenticated, request_id=req_abc123, user_id=12345")
	if err != nil {
		return
	}

	fmt.Println(log)
}
