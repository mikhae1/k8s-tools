///usr/bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const TIMEOUT = 3

type Res = struct {
	url  string
	code int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path as an argument (each line is an url to be checked)")
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

	timeout := time.Duration(TIMEOUT * time.Second)

	// Create a new http.Client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create a channel to receive response codes and URLs
	responseCodes := make(chan Res)

	// Launch a goroutine for each line
	for _, line := range lines {
		go func(line string) {
			url := line
			if !strings.HasPrefix(line, "http") {
				url = "https://" + line
			}

			resp, err := client.Get(url)
			var dnsErr *net.DNSError
			if err != nil {
				if errors.As(err, &dnsErr) {
					fmt.Printf("Can't resolve %s\n", line)
				} else {
					fmt.Printf("Error making GET request to %s: %s\n", url, err)
				}

				responseCodes <- Res{url, 0}
				return
			}

			defer resp.Body.Close()

			// Send the response code to the channel
			responseCodes <- Res{url, resp.StatusCode}
		}(line)
	}

	// Wait for responses from the channel
	var res []Res
	for i := 0; i < len(lines); i++ {
		res = append(res, <-responseCodes)
	}

	fmt.Printf("\nThe results:\n")
	for i := 0; i < len(res); i++ {
		fmt.Printf("-> %v\n", res[i])
	}
}
