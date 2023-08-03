///usr/bin/true; exec /usr/bin/env go run "$0" "$@"

//
// HTTPinger sends HTTP GET requests to a specified URL at regular intervals
// and measures the uptime and response status codes of the target server.
//
// Copyright (c) 2023 Mikhae1

package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

var URL = getArgEnv("URL", "http://example.com").(string)	// URL you want to ping
var DELAY = getArgEnv("DELAY", 500).(int)									// Delay between requests in ms
var TIMEOUT = getArgEnv("TIMEOUT", 5000).(int)						// Request timeout in ms
var BAUTH_USER = getArgEnv("BAUTH_USER", "").(string)			// HTTP Basic Auth BAUTH_USER
var BAUTH_PASS = getArgEnv("BAUTH_PASS", "").(string)			// HTTP Basic Auth BAUTH_PASS

type result struct {
	StatusCode int
	Err        error
	Timestamp  *time.Time
}

func main() {
	var wg sync.WaitGroup
	var results []result
	resultChan := make(chan result, 1) // Buffered channel with a capacity of 1
	stopChan := make(chan struct{})    // Signal channel to stop goroutines gracefully

	fmt.Printf("-> HTTPing %s with a delay of %vms (timeout %vms)...\n", URL, DELAY, TIMEOUT)

	timeout := time.Duration(TIMEOUT) * time.Millisecond

	go func() {
		for {
			select {
			case <-stopChan:
				// Stop the goroutine when receiving a signal on stopChan
				wg.Wait()
				close(resultChan)
				return
			default:
				wg.Add(1)
				go httPing(URL, timeout, resultChan, stopChan,  &wg)
				time.Sleep(time.Duration(DELAY) * time.Millisecond)
			}
		}
	}()

	// Stop on Ctrl-C
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	go func() {
		<-interruptChan
		fmt.Printf("\nReceived interrupt, waiting goroutines to finish (timeout=%v): ", timeout)
		// Signal the goroutine to stop by sending a signal on stopChan
		close(stopChan)
	}()

	// Collect results
	for r := range resultChan {
		results = append(results, r)
	}

	printResults(results)
}

func httPing(url string, timeout time.Duration, resultChan chan<- result, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	statusCode := 0
	status := "E"

	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		status = "X"

		fmt.Print(status)
		timestamp := time.Now()
		resultChan <- result{StatusCode: statusCode, Err: err, Timestamp: &timestamp}
		return
	}

	// Add HTTP Basic Auth headers
	if BAUTH_USER != "" && BAUTH_PASS != "" {
		authStr := fmt.Sprintf("%s:%s", BAUTH_USER, BAUTH_PASS)
		authEncoded := base64.StdEncoding.EncodeToString([]byte(authStr))
		req.Header.Set("Authorization", "Basic "+authEncoded)
	}

	select {
	case <-stopChan:
		// Immediately stop the goroutine if a stop signal is received
		return
	default:
		resp, err := client.Do(req)
		timestamp := time.Now()
		if err == nil {
			status = "."
			statusCode = resp.StatusCode
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				status = strconv.Itoa(resp.StatusCode)
			}
		}

		fmt.Print(status)
		resultChan <- result{StatusCode: statusCode, Err: err, Timestamp: &timestamp}
	}
}

func printResults(results []result) {
	successfulRequests := 0
	downtime := 0
	firstTimestamp := time.Time{} // Initialize firstTimestamp with zero value

	fmt.Printf("\nTotal Requests: %d\n", len(results))
	for _, res := range results {
		// Calculate the timestamp offset from the first request rounded to seconds
		timestampOffset := 0.0
		if res.Timestamp != nil && firstTimestamp.IsZero() {
			firstTimestamp = *res.Timestamp
		}
		if res.Timestamp != nil && !firstTimestamp.IsZero() {
			timestampOffset = res.Timestamp.Sub(firstTimestamp).Seconds()
		}

		fmt.Printf("\n[%6.1fs] Response Code: %d", timestampOffset, res.StatusCode)
		if res.Err != nil {
			fmt.Printf(", Error: %s", res.Err)
		}

		if res.StatusCode == http.StatusOK {
			successfulRequests++
		} else {
			downtime++
		}
	}
	fmt.Printf("\n\n")

	uptime := 100.0 - (float64(downtime) / float64(len(results)) * 100.0)
	totalDowntime := float64(downtime) * float64(DELAY) / 1000.0

	fmt.Printf("Successful Requests: %d\n", successfulRequests)
	fmt.Printf("Failed Requests: %d\n", downtime)
	fmt.Printf("Uptime: %.2f%%\n", uptime)
	fmt.Printf("Total Downtime: %.2f seconds\n", totalDowntime)
}

// GetEnvOrArg retrieves a setting from environment variables or command-line arguments,
// with a fallback to a default value.
//
// Parameters:
//   key      string      - The key representing the setting to be retrieved.
//   fallback interface{} - The fallback value to be used if the setting is not found.
//
// Returns:
//   interface{} - The value of the setting, converted to the appropriate type based on the 'fallback'.
//
// Copyright (c) 2023 Mikhae1
func getArgEnv(key string, fallback interface{}) interface{} {
	getValue := func(value string) interface{} {
		switch fallback.(type) {
		case string:
			return value
		case int:
			intVal, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("Error: Value '%s' must be an integer, but got '%s'\n", key, value)
				os.Exit(1)
			}
			return intVal
		default:
			fmt.Printf("Error: Unsupported type for fallback value of '%s'\n", key)
			os.Exit(1)
		}
		return nil
	}

	// Lookup in environment variables first
	if val, exists := os.LookupEnv(key); exists {
		return getValue(val)
	}

	// Search in os.Args
	for _, arg := range os.Args[1:] {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := parts[0]
		v := parts[1]

		if k == key {
			return getValue(v)
		}
	}

	// If not found, use fallback
	return fallback
}
