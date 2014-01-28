package grunk

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// custom io.WriteCloser to handle on the fly mp3 convertion

const FFMPEG = "ffmpeg"

type cmdWriter struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	aux   io.WriteCloser
}

func (w *cmdWriter) Write(p []byte) (n int, err error) {
	if w.aux != nil {
		w.aux.Write(p)
	}
	return w.stdin.Write(p)
}

func (w *cmdWriter) Close() error {
	w.stdin.Close()
	done := make(chan error)
	go func() {
		done <- w.cmd.Wait()
	}()
	select {
	case <-time.After(3 * time.Second):
		if err := w.cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill: ", err)
		}
		<-done // allow goroutine to exit
		log.Println("process killed")
	case err := <-done:
		log.Printf("process done with error = %v", err)
	}
	return w.cmd.Wait()
}

// This class will write the underlying stream (video) to a file
func getWriter(stream stream, name string) (out io.WriteCloser, err error) {
	var audio_bitrate uint
	path := cfg.OutputPath(stream, name)

	if _, err = os.Stat(path); err == nil && cfg.overwrite == false {
		return nil, fmt.Errorf("the destination file '%s' already exists and overwrite set to false", path)
	}

	audio_bitrate = AUDIO_BITRATE_MEDIUM
	for auto_audio_bitrate, qualities := range audioBitrates {
		for _, quality := range qualities {
			if quality == stream.Quality() {
				log.Printf("Auto-detected audio bitrate: '%dk'", auto_audio_bitrate)
				audio_bitrate = auto_audio_bitrate
				break
			}
		}
	}

	if cfg.isMp3() {
		fmt.Printf("Converting video to mp3 file at '%s' ...\n", path)
		out, err = getFFmpegWriter(path, audio_bitrate)
	} else {
		fmt.Printf("Downloading video to disk at '%s' ...\n", path)
		out, err = os.Create(path)
	}

	if err != nil {
		return nil, fmt.Errorf("opening destination file: %s", err)
	}

	log.Println("Destination opened at '%s'", path)

	return out, nil
}

func getDoubleWriter(stream stream, name string) (w *cmdWriter, err error) {
	writer1, err1 := getFFplayWriter()
	if err1 == nil {
		return nil, err1
	}
	writer2, err2 := getWriter(stream, name)
	if err2 == nil {
		return nil, err2
	}
	writer1.aux = writer2
	return writer1, nil
}

// This class streams to ffplay
func getFFplayWriter() (w *cmdWriter, err error) {
	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatalln("you need to install ffmpeg to run grunk")
		os.Exit(1)
	}

	w = &cmdWriter{
		exec.Command("/bin/sh", "-c", "ffmpeg -i - -f mp3 pipe:1 | mpg123 -"),
		nil,
		nil,
	}
	w.stdin, err = w.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	w.cmd.Start()
	return w, nil
}

// This class will convert and write to an mp3 file each stream
func getFFmpegWriter(path string, audio_bitrate uint) (w *cmdWriter, err error) {
	_, err = exec.LookPath(FFMPEG)
	if err != nil {
		return nil, fmt.Errorf("you need to install ffmpeg to convert to mp3: %s", err)
	}

	w = &cmdWriter{
		exec.Command(FFMPEG, "-i", "-", "-ab", fmt.Sprintf("%dk", audio_bitrate), path),
		nil,
		nil,
	}
	w.stdin, err = w.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	w.cmd.Start()
	return w, nil
}
