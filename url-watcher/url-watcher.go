///usr/bin/true; exec /usr/bin/env go run "$0" "$@"

/**
/* Watch given list of URLs
**/

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const TIMEOUT = 4
const LOG_PATH = "./logs"

type Res = struct {
	url  string
	code int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path as an argument (each line is an url to be checked)")
		os.Exit(1)
	}

	fname := os.Args[1]

	file, err := os.Open(fname)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if len(LOG_PATH) > 0 {
		defer initLogger(LOG_PATH, fname).Close()
	}

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

			var dnsErr *net.DNSError
			resp, err := client.Get(url)
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
			responseCodes <- Res{url, resp.StatusCode}
		}(line)
	}

	// Wait for responses from the channel
	var res []Res
	for i := 0; i < len(lines); i++ {
		res = append(res, <-responseCodes)
	}

	sort.Slice(res, func(i, j int) bool { return res[i].code < res[j].code })
	log.Printf("\nAll responses from %v urls:\n", len(res))
	for i, r := range res {
		log.Printf("%4v. %v, response: %v \n", i+1, r.url, r.code)
	}

	res200 := filterByCode(res, 200)
	log.Printf("\nUrls with response code 200 (%v urls):\n", len(res200))
	for i, r := range res200 {
		log.Printf("%4v. %v, response: %v \n", i+1, r.url, r.code)
	}

}

func filterByCode(res []Res, code int) []Res {
	var filteredRes []Res
	for _, r := range res {
		if r.code == code {
			filteredRes = append(filteredRes, r)
		}
	}
	return filteredRes
}

func initLogger(logPath string, fname string) *os.File {
	currentTime := time.Now().Format("01-02-06_15-04-05")

	logFile, err := os.Create(logPath + "/" + strings.ReplaceAll(fname, "/", "-") + "-" + currentTime + ".log")
	if err != nil {
		log.Fatal(err)
	}
	log.SetFlags(0)

	// redirect standard output to the log file
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return logFile
}
