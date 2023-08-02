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

var URL = getEnv("URL", "http://example.com") // URL you want to ping
var DELAY = getEnv("DELAY", "500")            // Delay between requests in ms
var TIMEOUT = getEnv("TIMEOUT", "5000")				// Request timeout in ms
var BAUTH_USER = getEnv("BAUTH_USER", "")     // HTTP Basic Auth BAUTH_USER
var BAUTH_PASS = getEnv("BAUTH_PASS", "")     // HTTP Basic Auth BAUTH_PASS

type result struct {
	StatusCode int
	Err        error
}

func main() {

	resultChan := make(chan result, 1) // Buffered channel with a capacity of 1
	var wg sync.WaitGroup
	var statuses []result

	delay, err := strconv.Atoi(DELAY)
	if err != nil {
		fmt.Println("You should set DELAY to integer:", err)
		return
	}

	fmt.Printf("Pinging %s with a delay of %vms...\n", URL, delay)

	timeout := time.Duration(3*delay) * time.Millisecond
	if TIMEOUT != "" {
		t, err := strconv.Atoi(TIMEOUT)
		if err != nil {
			fmt.Println("You should set DELAY to integer:", err)
			return
		}
		timeout = time.Duration(t) * time.Millisecond
	}

	go func() {
		for {
			wg.Add(1)
			go pingURL(URL, resultChan, &wg, timeout)
			time.Sleep(time.Duration(delay) * time.Millisecond)
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
	delayInt, err := strconv.Atoi(DELAY)
	if err != nil {
		fmt.Println("Error converting DELAY to integer:", err)
		return
	}
	totalDowntime := float64(downtime) * float64(delayInt) / 1000.0 // Convert DELAY to integer before converting to float64
	fmt.Printf("Successful Requests: %d\n", successfulRequests)
	fmt.Printf("Failed Requests: %d\n", downtime)
	fmt.Printf("Uptime: %.2f%%\n", uptime)
	fmt.Printf("Total Downtime: %.2f seconds\n", totalDowntime)
}

func getEnv(key, fallback string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		val = fallback
	}
	return val
}
