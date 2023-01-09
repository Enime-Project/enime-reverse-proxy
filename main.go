package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func main() {
	client := &http.Client{}

	originServerHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rawUrl := req.URL.Query().Get("url")
		if rawUrl == "" {
			_, _ = fmt.Fprint(rw, "You need to specify an URL to proxy with.")
			return
		}

		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			_, _ = fmt.Fprint(rw, "The URL needs to be encoded beforehand.")
			return
		}

		req.URL = parsedUrl

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(body))

		proxyReq, err := http.NewRequest(req.Method, parsedUrl.String(), bytes.NewReader(body))

		proxyReq.Header = make(http.Header)
		for h, val := range req.Header {
			proxyReq.Header[h] = val
		}

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadGateway)
			return
		}

		rw.Header().Set("Access-Control-Allow-Origin", "*")
		for k, v := range resp.Header {
			if k == "Access-Control-Allow-Origin" {
				continue
			}
			for _, s := range v {
				rw.Header().Add(k, s)
			}
		}

		defer resp.Body.Close()

		rw.WriteHeader(resp.StatusCode)
		io.Copy(rw, resp.Body)
		resp.Body.Close()
	})

	log.Fatal(http.ListenAndServe(":80", originServerHandler))
}
