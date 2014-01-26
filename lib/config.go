package grunk

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	// "path/filepath"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
	"strings"
)

var audioBitrates map[uint][]string = map[uint][]string{
	AUDIO_BITRATE_LOW:    []string{QUALITY_SMALL},
	AUDIO_BITRATE_MEDIUM: []string{QUALITY_MEDIUM, QUALITY_LARGE, QUALITY_UNKNOWN},
	AUDIO_BITRATE_HIGH:   []string{QUALITY_HD720, QUALITY_HD1080, QUALITY_HIGHRES},
}

var sortedQualities []string = []string{
	QUALITY_SMALL,
	QUALITY_MEDIUM,
	QUALITY_LARGE,
	QUALITY_HD720,
	QUALITY_HD1080,
	QUALITY_HIGHRES,
	QUALITY_UNKNOWN,
}

var formatsTrigger map[string]string = map[string]string{
	FORMAT_MP4:  "video/mp4",
	FORMAT_FLV:  "video/x-flv",
	FORMAT_WEBM: "video/webm",
	FORMAT_3GP:  "video/3gpp",
}

var sortedFormats []string = []string{
	FORMAT_MP4,
	FORMAT_FLV,
	FORMAT_WEBM,
	FORMAT_3GP,
	FORMAT_UNKNOWN,
}

// comma delimited parameters

type commaStringList struct {
	values  []string
	allowed map[string]struct{}
}

func (sl *commaStringList) String() string {
	return strings.Join(sl.values, ",")
}

func (sl *commaStringList) Set(value string) error {
	sl.values = sl.values[0:0]
	var exists bool
	for _, s := range strings.Split(value, ",") {
		_, exists = sl.allowed[s]
		if len(sl.allowed) > 0 && !exists {
			return errors.New(fmt.Sprintf("non allowed value '%s'", s))
		}
		sl.values = append(sl.values, s)
	}
	return nil
}

func CreateCommaStringList(values []string, allowed []string) *commaStringList {
	sl := &commaStringList{[]string{}, map[string]struct{}{}}
	for _, value := range values {
		sl.values = append(sl.values, value)
	}
	for _, value := range allowed {
		sl.allowed[value] = struct{}{}
	}
	return sl
}

// our config struct

type Config struct {
	verbose      bool
	output       string // path
	overwrite    bool
	quality      *commaStringList
	format       *commaStringList
	videoId      string
	toMp3        bool
	audioBitrate uint
}

var cfg *Config = &Config{
	false,
	DEFAULT_DESTINATION,
	false,
	CreateCommaStringList(
		sortedQualities,
		append([]string{QUALITY_MAX, QUALITY_MIN}, sortedQualities...),
	),
	CreateCommaStringList(
		[]string{FORMAT_MP4, FORMAT_FLV, FORMAT_WEBM, FORMAT_3GP},
		sortedFormats,
	),
	"",
	false,
	AUDIO_BITRATE_AUTO,
}

// reads the videoId property and try to find what we need inside
func (cfg *Config) findVideoId() (videoId string, err error) {
	videoId = cfg.videoId
	if strings.Contains(videoId, "youtu") || strings.ContainsAny(videoId, "\"?&/<%=") {
		// log.Println("Provided video id seems to be an url, trying to detect")
		re_list := []*regexp.Regexp{
			regexp.MustCompile(`(?:v|embed|watch\?v)(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`([^"&?/=%]{11})`),
		}
		for _, re := range re_list {
			if is_match := re.MatchString(videoId); is_match {
				subs := re.FindStringSubmatch(videoId)
				videoId = subs[1]
			}
		}
	}
	// log.Printf("Found video id: '%s'", videoId)
	if strings.ContainsAny(videoId, "?&/<%=") {
		return videoId, errors.New("invalid characters in video id")
	}
	if len(videoId) < 10 {
		return videoId, errors.New("the video id must be at least 10 characters long")
	}
	return videoId, nil
}

func (cfg *Config) isMp3() bool {
	return cfg.toMp3
}

func (cfg *Config) OutputPath(stream stream, name string) (path string) {
	if stream != nil {
		path = strings.Replace(cfg.output, "%format%", stream.Format(), -1)
		path = strings.Replace(path, "%title%", stream["title"], -1)
		path = strings.Replace(path, "%author%", stream["author"], -1)
	} else {
		path = strings.Replace(DEFAULT_DESTINATION_MP3, "%title%", name, -1)
	}
	return
}

func (cfg *Config) AudioBitrate(stream stream) (audio_bitrate uint) {
	if cfg.audioBitrate != AUDIO_BITRATE_AUTO {
		log.Printf("Manually set audio bitrate: '%dk'", cfg.audioBitrate)
		return cfg.audioBitrate
	}
	for audio_bitrate, qualities := range audioBitrates {
		for _, quality := range qualities {
			if quality == stream.Quality() {
				log.Printf("Auto-detected audio bitrate: '%dk'", audio_bitrate)
				return audio_bitrate
			}
		}
	}

	audio_bitrate = AUDIO_BITRATE_MEDIUM
	log.Printf("WARNING: not bitrate defined for this quality '%s', defaulting to '%dk'", stream.Quality(), audio_bitrate)
	return
}

func (cfg *Config) selectStream(streams streamList) (stream stream, err error) {
	if len(streams) < 1 {
		return nil, errors.New("no streams found")
	}
	valid_streams := streamList{}
	for _, format := range cfg.format.values {
		for _, s := range streams {
			if s.Format() == format {
				valid_streams = append(valid_streams, s)
			}
		}
		if len(valid_streams) >= 1 {
			// log.Printf("Found format '%s', with %d streams", format, len(valid_streams))
			break
		}
	}
	if len(valid_streams) < 1 {
		return nil, errors.New("no streams match the requested formats")
	}
	streams = valid_streams
	valid_streams = streamList{}
	for _, quality := range cfg.quality.values {
		// log.Printf("quality %s", quality)
		for _, s := range streams {
			if s.Quality() == quality {
				valid_streams = append(valid_streams, s)
			} else {
				// log.Printf("Rejecting Stream with quality '%s'", s.Quality())
			}
		}
		if len(valid_streams) >= 1 {
			// log.Printf("Found quality '%s', with %d streams", quality, len(valid_streams))
			break
		}
	}
	if len(valid_streams) < 1 {
		return nil, errors.New("no streams match the requested qualities")
	}
	return valid_streams[0], nil
}

func CookieGrabber() string {
	// Pick up a plug.dj cookie from the same directory that we're in
	// Format is cookie.dat
	// exit if we fail
	content, err := ioutil.ReadFile("cookie.dat")
	if err != nil {
		log.Println("To access plug.dj APIs, we need a Cookie to authenticate")
		log.Println("This couldn't be found, so we're going to attempt to recover it from your browser (requires you having logged in once)")
		res, err := attemptCookieSteal()
		if err != nil {
			log.Println("fatal error in cookie stealing")
			os.Exit(1)
		}
		return res

	}
	return string(content)
}

func attemptCookieSteal() (string, error) {
	var loc string
	switch runtime.GOOS {
	case "darwin":
		loc = os.Getenv("HOME") + "/Library/Application Support/Google/Chrome/Default/Cookies"
	case "linux":
		loc = os.Getenv("HOME") + "/.config/google-chrome/Default/Cookies"
	default:
		log.Println("no idea what OS", runtime.GOOS, "is")
		return "", nil
	}
	log.Println("trying location", loc)
	result, err := cookieStealer(loc)
	if err != nil {
		log.Fatal("Error getting cookie.")
	} else {
		log.Println("Got value", result)
	}
	return result, err
}

func cookieStealer(cookieLoc string) (string, error) {
	db, err := sql.Open("sqlite3", cookieLoc)
	if err != nil {

		log.Fatal(err)
		return "", err
	}
	defer db.Close()

	sql := `
	SELECT host_key, path, secure, expires_utc, name, value FROM cookies where host_key like '%plug%'
	`
	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		// log.Println(r)
		var host_key string
		var path string
		var secure int
		var expires_utc int
		var name string
		var value string
		rows.Scan(&host_key, &path, &secure, &expires_utc, &name, &value)
		log.Println("host_key:", host_key, "path:", path, "secure:", secure, "expires_utc:", expires_utc, "Name:", name, "value:", value)
		if name == "usr" {
			return "usr=" + value, nil
		}
	}
	rows.Close()
	return "", errors.New("USR NOT FOUND")
}
