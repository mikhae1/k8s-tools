# k8s-tools

Everyday tools for Kubernetes-related activities.

## HTTPinger
HTTPinger is a command-line tool written in Go that sends HTTP GET requests to a specified URL at regular intervals. It measures the uptime and response status codes of the target server.

The tool operates asynchronously, meaning it can perform multiple requests concurrently without waiting for each request to complete before starting the next one.

### Usage
```sh
$ URL=http://example.com DELAY=300 ./httping.go
```

### Configuration
The following environment variables can be used to configure HTTPinger:

    URL: The URL you want to ping. Defaults to http://example.com.
    DELAY: Delay between requests in milliseconds. Defaults to 500ms.
    TIMEOUT: Request timeout in milliseconds. Defaults to 5000ms.
    BAUTH_USER: HTTP Basic Auth username (optional).
    BAUTH_PASS: HTTP Basic Auth password (optional).


### Output
```sh
$ URL=https://example.com DELAY=300 TIMEOUT=500 ./httping.go

-> HTTPing https://example.com with a delay of 300ms (timeout 600ms)...
E........^C
Received interrupt, waiting goroutines to finish (timeout=600ms):
Total Requests: 9

[   0.0s] Response Code: 0, Error: Get "https://example.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
[   0.1s] Response Code: 200
[   0.2s] Response Code: 200
[   0.5s] Response Code: 200
[   0.8s] Response Code: 200
[   1.1s] Response Code: 200
[   1.4s] Response Code: 200
[   1.7s] Response Code: 200
[   2.0s] Response Code: 200

Successful Requests: 8
Failed Requests: 1
Uptime: 88.89%
Total Downtime: 0.30 seconds
```

**Where:**
- Total Requests: The total number of requests made.
Successful Requests: The number of requests with a successful HTTP response (status code 200).
- Failed Requests: The number of requests with non-successful HTTP responses (status code other than 200).
- Uptime: The percentage of successful requests out of the total requests, indicating the server's uptime.
- Total Downtime: The total duration of downtime, calculated based on the number of failed requests and the delay between requests.


## url-watcher

This Go program reads a file containing a list of URLs, makes a parallel GET requests to each URL, and outputs the response codes for them.

### Usage

    url_checker.go urls.txt

This will read the file urls.txt and make GET requests to each URL in the file. The output will be in the following format:

    Response codes with 200 code (2 urls):
        1. https://example.com, response: 200
        1. https://your-app.k8s.local, response: 200

### Additional dependencies

This program uses the core packages only.
