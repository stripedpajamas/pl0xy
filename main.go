package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var httpClient = http.Client{}
var targetURL string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pl0xy <target>")
		os.Exit(1)
	}
	targetURL = os.Args[1]
	if targetURL[len(targetURL)-1] == '/' {
		// strip final slash if it's there
		targetURL = targetURL[:len(targetURL)-1]
	}

	// set up routing
	http.HandleFunc("/", pl0xyHandler)
	http.ListenAndServe(":3001", nil)
}

func pl0xyHandler(w http.ResponseWriter, r *http.Request) {
	// get route, query, headers
	route := r.URL.Path
	query := r.URL.RawQuery
	headers := r.Header

	// generate target url
	url := fmt.Sprintf("%s%s?%s", targetURL, route, query)

	// send request to target
	res, err := func() (*http.Response, error) {
		req, err := http.NewRequest(r.Method, url, r.Body)
		if err != nil {
			return nil, err
		}

		// set the headers
		req.Header = headers

		// make the request
		return httpClient.Do(req)
	}()

	// if the request didn't work, send back a 500
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send back the response
	for k, v := range res.Header {
		w.Header().Set(k, strings.Join(v, ""))
	}
	w.WriteHeader(res.StatusCode)
	w.Write(response)
}
