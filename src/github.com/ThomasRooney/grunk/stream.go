package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type stream map[string]string
type streamList []stream

func (s stream) Url() string {
	return s["url"] + "&signature=" + s["sig"]
}

func (s stream) Format() string {
	for format, trigger := range formatsTrigger {
		if strings.Contains(s["type"], trigger) {
			return format
		}
	}
	return FORMAT_UNKNOWN
}

func (s stream) Quality() string {
	for _, quality := range sortedQualities {
		if quality == s["quality"] {
			return quality
		}
	}
	return QUALITY_UNKNOWN
}

func downloadFromUrl(url string, out io.Writer) error {

	log.Printf("Running stream from '%s'", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("requesting stream: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("reading answer: non 200 status code received: '%s'", err)
	}
	length, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("saving file: %s (%d bytes copied)", err, length)
	}

	log.Printf("Downloaded %d bytes", length)
	return nil
}

func (stream stream) download(out io.Writer) error {
	url := stream.Url()
	return downloadFromUrl(url, out)
}
