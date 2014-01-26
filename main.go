package main

import (
	grunk "github.com/ThomasRooney/grunk/lib"
	"log"
	"strings"
	"time"
)

func main() {
	// log.SetOutput(ioutil.Discard)
	was_error := false
	AUTH_COOKIE := grunk.CookieGrabber()
	for {
		if was_error {
			// log.Println("waiting 10 seconds")
			time.Sleep(10 * time.Second)
			was_error = false
			// log.Println("done")
		}
		media := grunk.GetRoomMedia("tastycat", AUTH_COOKIE)
		for k, v := range media {
			log.Println("Key: ", k, "Value: ", v)
		}

		toks := strings.Split(media["id"], ":")
		switch toks[0] {
		case "1": // youtube
			was_error = !grunk.PlayYoutube(toks[1])
			log.Println("back in main")
		case "2": // soundcloud
			was_error = !grunk.PlaySoundcloud(toks[1])
		default:
			was_error = true
			log.Printf("id: %s, Grunk can't handle streams of this form yet", toks[0])
		}
		// log.Println("waiting 1 second")
		time.Sleep(time.Second)
		// log.Println("done")
	}

}
