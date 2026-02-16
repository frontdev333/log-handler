package main

import (
	"fmt"
	"log-handler/internal/logentry"
	"path"
)

func main() {
	filepath := "./examples/test.log"
	res, err := logentry.ReadLogFile(filepath)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		fmt.Println("===ERROR===")
	}

	fmt.Println("Processing file:", path.Base(filepath))
	fmt.Println("Total lines:", len(res))

}
