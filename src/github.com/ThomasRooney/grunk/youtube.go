package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func getVideoInfo(videoId string) (string, error) {
	url := "http://youtube.com/get_video_info?video_id=" + videoId
	log.Printf("Requesting url: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("An error occured while requesting the video information: '%s'", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("An error occured while requesting the video information: non 200 status code received: '%s'", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("An error occured while reading the video information: '%s'", err)
	}
	log.Printf("Got %d bytes answer", len(body))
	return string(body), nil
}

func decodeVideoInfo(response string) (streams streamList, err error) {
	// decode

	answer, err := url.ParseQuery(response)
	if err != nil {
		err = fmt.Errorf("parsing the server's answer: '%s'", err)
		return
	}

	// check the status

	status, ok := answer["status"]
	if !ok {
		err = fmt.Errorf("no response status found in the server's answer")
		return
	}
	if status[0] == "fail" {
		reason, ok := answer["reason"]
		if ok {
			err = fmt.Errorf("'fail' response status found in the server's answer, reason: '%s'", reason[0])
		} else {
			err = errors.New(fmt.Sprint("'fail' response status found in the server's answer, no reason given"))
		}
		return
	}
	if status[0] != "ok" {
		err = fmt.Errorf("non-success response status found in the server's answer (status: '%s')", status)
		return
	}

	// log.Printf("Server answered with a success code")

	/*
	   for k, v := range answer {
	           log.Printf("%s: %#v", k, v)
	   }
	*/

	// read the streams map

	stream_map, ok := answer["url_encoded_fmt_stream_map"]
	if !ok {
		err = errors.New(fmt.Sprint("no stream map found in the server's answer"))
		return
	}

	// read each stream

	streams_list := strings.Split(stream_map[0], ",")

	// log.Printf("Found %d streams in answer", len(streams_list))

	for stream_pos, stream_raw := range streams_list {
		stream_qry, err := url.ParseQuery(stream_raw)
		if err != nil {
			log.Printf(fmt.Sprintf("An error occured while decoding one of the video's stream's information: stream %d: %s\n", stream_pos, err))
			continue
		}
		stream := stream{
			"quality": stream_qry["quality"][0],
			"type":    stream_qry["type"][0],
			"url":     stream_qry["url"][0],
			"sig":     stream_qry["sig"][0],
			"title":   answer["title"][0],
			"author":  answer["author"][0],
		}
		streams = append(streams, stream)
		// if quality == QUALITY_UNKNOWN {
		// log.Printf("Found unknown quality '%s'", stream["quality"])
		// }

		// format := stream.Format()
		// if format == FORMAT_UNKNOWN {
		// log.Printf("Found unknown format '%s'", stream["type"])
		// }

		// log.Printf("Stream found: quality '%s', format '%s'", quality, format)
	}

	// log.Printf("Successfully decoded %d streams", len(streams))

	return
}
