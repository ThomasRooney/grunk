package grunk

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const SOUNDCLOUD_ID = "27b079ec70d5787cee129f9b31ba5f4f"

func constructCookie(name string, value string) *http.Cookie {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Path = "/"
	cookie.HttpOnly = false
	cookie.Secure = false
	return cookie
}

func GetRoomMedia(room string, AUTH_COOKIE string) map[string]string {
	url := "http://plug.dj/_/gateway/room.details"
	request := `{"service":"room.details","body":["` + room + `"]}`

	client := &http.Client{}

	wrapped_json := strings.NewReader(request)
	// client := http.Client{Jar: jar}
	req, err := http.NewRequest("POST", url, wrapped_json)
	if err != nil {
		log.Println("NewRequest Error")
		log.Fatal(err)
	}
	req.Header.Set("Cookie", AUTH_COOKIE)
	req.Header.Set("Origin", `http://plug.dj`)
	req.Header.Set("Accept-Encoding", `gzip,deflate,sdch`)
	// req.Header.Set("Accept-Encoding", `application/json`)
	req.Header.Set("Host", `plug.dj`)
	req.Header.Set("Accept-Language", `en-US,en;q=0.8`)
	req.Header.Set("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.63 Safari/537.36`)
	req.Header.Set("Content-Type", `application/json`)
	req.Header.Set("Accept", `application/json, text/javascript, */*; q=0.01`)
	req.Header.Set("Referer", `http://plug.dj/tastycat/`)
	req.Header.Set("X-Requested-With", `XMLHttpRequest`)
	req.Header.Set("Connection", `keep-alive`)
	resp, err := client.Do(req)
	// (url, "application/json", wrapped_json)
	if err != nil {
		// Handle err
		log.Println("client.Do error: %s", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Decode the ascii reader to a json

	u := make(map[string]interface{})

	decoder := json.NewDecoder(reader)

	decoder.Decode(&u)
	b := u["body"].(map[string]interface{})
	r := b["room"].(map[string]interface{})
	m := r["media"].(map[string]interface{})

	media := make(map[string]string)
	for k, v := range m {
		media[k], _ = v.(string) // strip out all non-string formats
	}
	return media
}

func getCmdWriter(s stream, filename string) (*cmdWriter, error) {
	return getFFplayWriter()
}

func PlayYoutube(id string) (success bool) {
	success = false
	response, err := getVideoInfo(id)
	streams, err := decodeVideoInfo(response)
	stream, err := cfg.selectStream(streams)
	if err != nil {
		log.Printf("ERROR: unable to select a stream: %s\n", err)
		return
	}

	out, err := getCmdWriter(stream, "")
	if err != nil {
		log.Printf("ERROR: unable to create the output writer: %s\n", err)
		return
	}
	defer func() {
		log.Println("Closing pipe")
		err = out.Close()
		log.Println("Closed")
		if err != nil {
			log.Printf("ERROR: unable to close destination: %s\n", err)
			return
		}
	}()
	err = stream.download(out)

	if err != nil {
		log.Printf("ERROR: unable to download the stream: %s\n", err)
		return
	}
	log.Println("successful `stream play.. returning")
	success = true
	return
}

func PlaySoundcloud(id string) (success bool) {
	success = false

	details_url := "http://api.soundcloud.com/tracks/" + id + ".json?client_id=" + SOUNDCLOUD_ID
	req, err := http.NewRequest("GET", details_url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	reader := resp.Body
	u := make(map[string]interface{})

	decoder := json.NewDecoder(reader)

	decoder.Decode(&u)
	log.Println("soundcloud-details")
	for k, v := range u {
		log.Println("Key: ", k, "Val: ", v)
	}
	stream_url := u["stream_url"].(string) + "?client_id=" + SOUNDCLOUD_ID
	out, err := getCmdWriter(nil, "")
	if err != nil {
		log.Printf("ERROR: unable to create the output writer: %s\n", err)
		return
	}
	defer func() {
		err = out.Close()
		if err != nil {
			log.Printf("ERROR: unable to close destination: %s\n", err)
			return
		}
	}()
	downloadFromUrl(stream_url, out)

	success = true
	return
}
