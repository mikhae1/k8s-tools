# URL Checker

This Go program reads a file containing a list of URLs and makes a GET request to each URL, and outputs the response code and the url.

## Usage

    url_checker.go urls.txt

This will read the file urls.txt and make GET requests to each URL in the file. The output will be in the following format:

    Response code: 200
    Response code: 404

## Note

Please note that this sample code is intended to serve as an example only and may require additional error handling or other modifications to suit your specific needs.

## Additional dependencies

This program uses the core packages only.
