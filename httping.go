///usr/bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var URL = getEnv("URL", "http://example.com").(string)	// URL you want to ping
var DELAY = getEnv("DELAY", 500).(int)									// Delay between requests in ms
var TIMEOUT = getEnv("TIMEOUT", 5000).(int)							// Request timeout in ms
var BAUTH_USER = getEnv("BAUTH_USER", "").(string)			// HTTP Basic Auth BAUTH_USER
var BAUTH_PASS = getEnv("BAUTH_PASS", "").(string)			// HTTP Basic Auth BAUTH_PASS

type result struct {
	StatusCode int
	Err        error
}

func main() {
	resultChan := make(chan result, 1) // Buffered channel with a capacity of 1
	var wg sync.WaitGroup
	var statuses []result

	fmt.Printf("Pinging %s with a delay of %vms (timeout %vms)...\n", URL, DELAY, TIMEOUT)

	timeout := time.Duration(TIMEOUT) * time.Millisecond

	go func() {
		for {
			wg.Add(1)
			go pingURL(URL, resultChan, &wg, timeout)
			time.Sleep(time.Duration(DELAY) * time.Millisecond)
		}
	}()

	// stop on Ctrl-C
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interruptChan
		fmt.Println("\nReceived interrupt, waiting for goroutines to finish...")
		wg.Wait()        // Wait for all goroutines to finish
		close(resultChan) // Close the resultChan after all goroutines have finished
	}()

Loop:
	for {
		select {
		case status, ok := <-resultChan:
			if !ok {
				break Loop
			}
			statuses = append(statuses, status)
		}
	}

	printResults(statuses)
}

func pingURL(url string, resultChan chan<- result, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		status := "E" // Error status code
		fmt.Print(status)
		resultChan <- result{StatusCode: 0, Err: err}
		return
	}

	if BAUTH_USER != "" && BAUTH_PASS != "" {
		// Add Basic Auth headers
		authStr := fmt.Sprintf("%s:%s", BAUTH_USER, BAUTH_PASS)
		authEncoded := base64.StdEncoding.EncodeToString([]byte(authStr))
		req.Header.Set("Authorization", "Basic "+authEncoded)
	}

	status := "."
	resp, err := client.Do(req)
	if err != nil {
		status = "E" // Error status code
		fmt.Print(status)
		resultChan <- result{StatusCode: 0, Err: err}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status = strconv.Itoa(resp.StatusCode)
	}
	fmt.Print(status)

	resultChan <- result{StatusCode: resp.StatusCode, Err: nil}
}

func printResults(statuses []result) {
	fmt.Printf("\nResulting Table:\n")
	fmt.Printf("Total Requests: %d\n", len(statuses))
	successfulRequests := 0
	downtime := 0
	for _, res := range statuses {
		fmt.Print("\nResponse Code: ", res.StatusCode)
		if res.Err != nil {
			fmt.Print(", ", res.Err)
		}

		if res.StatusCode == http.StatusOK {
			successfulRequests++
		} else {
			downtime++
		}
	}
	fmt.Printf("\n\n")

	uptime := 100.0 - (float64(downtime) / float64(len(statuses)) * 100.0)
	totalDowntime := float64(downtime) * float64(DELAY) / 1000.0

	fmt.Printf("Successful Requests: %d\n", successfulRequests)
	fmt.Printf("Failed Requests: %d\n", downtime)
	fmt.Printf("Uptime: %.2f%%\n", uptime)
	fmt.Printf("Total Downtime: %.2f seconds\n", totalDowntime)
}

func getEnv(key string, fallback interface{}) interface{} {
	val, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	// Check the type of the fallback value to convert the environment variable accordingly
	switch fallback.(type) {
	case string:
		return val
	case int:
		intVal, err := strconv.Atoi(val)
		if err != nil {
			fmt.Printf("Error: Environment variable '%s' must be an integer, but got '%s'\n", key, val)
			os.Exit(1)
		}
		return intVal
	default:
		fmt.Printf("Error: Unsupported type for fallback value of environment variable '%s'\n", key)
		os.Exit(1)
	}

	return nil
}
