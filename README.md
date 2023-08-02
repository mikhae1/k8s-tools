# k8s-tools

Everyday tools for Kubernetes-related activities

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
