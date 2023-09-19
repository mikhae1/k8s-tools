# k8s-tools

Everyday tools for Kubernetes-related activities.
- [k8s-tools](#k8s-tools)
  - [HTTPing](#httping)
    - [Usage](#usage)
    - [Output](#output)
    - [Configuration](#configuration)
  - [url-watcher](#url-watcher)
    - [Usage](#usage-1)
    - [Additional dependencies](#additional-dependencies)
  - [eks-top](#eks-top)
    - [Output](#output-1)
    - [Prerequisites](#prerequisites)
    - [Usage](#usage-2)

## HTTPing
HTTPing is a command-line tool written in Go that sends HTTP GET requests to a specified URL at millisecond intervals. It measures the uptime and response status codes of the target server.

The tool operates asynchronously, meaning it can perform multiple requests concurrently without waiting for each request to complete before starting the next one.

### Usage
```sh
$ URL=http://example.com DELAY=300 ./httping.go
```

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


### Configuration
The following environment variables can be used to configure HTTPing:

- **URL**: The URL you want to ping. Defaults to http://example.com.
- **DELAY**: Delay between requests in milliseconds. Defaults to 500ms.
- **TIMEOUT**: Request timeout in milliseconds. Defaults to 5000ms.
- **BAUTH_USER**: HTTP Basic Auth username (optional).
- **BAUTH_PASS**: HTTP Basic Auth password (optional).


## url-watcher

This Go tool reads a file containing a list of URLs, makes a parallel GET requests to each URL, and outputs the response codes for them.

### Usage

    url-watcher/url-watcher.go urls.txt

This will read the file urls.txt and make GET requests to each URL in the file. The output will be in the following format:

    Response codes with 200 code (2 urls):
        1. https://example.com, response: 200
        1. https://your-app.k8s.local, response: 200

### Additional dependencies

This tool uses the core packages only.

## eks-top
This Python script is designed to monitor AWS EKS clusters and their associated EC2 instances, providing insights into resource utilization such as CPU, memory, ephemeral storage, and disk usage. It uses various AWS and Kubernetes APIs to collect and display this information in a tabular format.

### Output

```sh
+---------------+--------------+--------------+----------------------+------------+------------+------------+---------+----------+-----------+------+
| Instance Name | Instance Type| Private IP   | Instance ID          | CPU Avg(1d)| CPU Avg(7d)| Max CPU(7d)| Mem Util| Eph.Usage| Disk Util | AGE  |
+---------------+--------------+--------------+----------------------+------------+------------+------------+---------+----------+-----------+------+
| worker1       | t2.micro     | 10.0.0.101   | i-1234567890abcdef01 | 22.5%      | 23.1%      | 24.7%      | 45.6%   | No Data  | 19.8%     | 15d  |
| worker2       | m5.large     | 10.0.0.102   | i-234567890abcdef12  | 12.7%      | 14.3%      | 15.9%      | 63.2%   | 28.5%    | 82.1%     | 31d  |
| worker3       | c5.xlarge    | 10.0.0.103   | i-34567890abcdef23   | 45.8%      | 47.2%      | 49.6%      | 78.9%   | 54.2%    | 12.5%     | 7d   |
| worker4       | t3.micro     | 10.0.0.104   | i-4567890abcdef3456  | No Data    | No Data    | No Data    | 32.1%   | 16.7%    | 42.8%     | 54d  |
| worker5       | m4.large     | 10.0.0.105   | i-567890abcdef4567   | 34.2%      | 36.8%      | 38.4%      | 51.3%   | 42.0%    | No Data   | 23d  |
+---------------+--------------+--------------+----------------------+------------+------------+------------+---------+----------+-----------+------+
```

### Prerequisites
Before using this script, ensure that you have the following prerequisites:

- Python 3 installed on your system.

The necessary Python packages installed. You can install them using the following command:

    pip3 install -r ./eks-top/requirements.txt


AWS CLI configured with the necessary credentials and region. You can configure AWS CLI using `aws configure` command.

### Usage

    ./eks-top/eks-top.py
