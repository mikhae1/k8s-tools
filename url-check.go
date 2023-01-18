package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path as an argument")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	// Create a channel to receive response codes
	responseCodes := make(chan int)

	// Launch a goroutine for each line
	for _, line := range lines {
		go func(line string) {
			if !strings.HasPrefix(line, "http") {
				line = "https://" + line
			}

			resp, err := http.Get(line)
			if err != nil {
				fmt.Printf("Error making GET request to %s: %s\n", line, err)

				responseCodes <- 0
				return
			}

			defer resp.Body.Close()

			// Send the response code to the channel
			responseCodes <- resp.StatusCode
		}(line)
	}

	// Receive the response codes from the channel
	for i := 0; i < len(lines); i++ {
		fmt.Printf("Response code: %d\n", <-responseCodes)
	}
}
