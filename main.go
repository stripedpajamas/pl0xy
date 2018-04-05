package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var httpClient = http.Client{}
var targetURL string

func main() {
	if len(os.Args) < 2 {
		printUsage()
	}

	targetURL = os.Args[1]
	if targetURL[len(targetURL)-1] == '/' {
		// strip final slash if it's there
		targetURL = targetURL[:len(targetURL)-1]
	}

	port := 4444
	if len(os.Args) > 2 {
		parsedArg, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Error parsing port argument")
			printUsage()
		}
		port = parsedArg
	}

	log.Printf("Target: %s", targetURL)
	log.Printf("Listening for requests on port %d", port)

	// set up routing
	http.HandleFunc("/", pl0xyHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func printUsage() {
	fmt.Println("Usage: pl0xy <target> [port (default 4444)]")
	os.Exit(1)
}

func pl0xyHandler(w http.ResponseWriter, r *http.Request) {
	// generate target url
	url := fmt.Sprintf("%s%s?%s", targetURL, r.URL.Path, r.URL.RawQuery)

	log.Printf("%s %s Content-Length: %d", r.Method, url, r.ContentLength)

	// send request to target
	res, err := func() (*http.Response, error) {
		req, err := http.NewRequest(r.Method, url, r.Body)
		if err != nil {
			return nil, err
		}

		// set the headers
		req.Header = r.Header

		// make the request
		return httpClient.Do(req)
	}()

	// if the request didn't work, send back a 500
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set headers on response to headers we received
	for k, v := range res.Header {
		w.Header().Set(k, v[0])
	}

	// set status code to status code we received
	w.WriteHeader(res.StatusCode)

	// write body to caller
	_, err = io.Copy(w, res.Body)
	defer res.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
