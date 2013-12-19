package main

import (
	"fmt"
	"io"
	"time"

	"log"
	"os"
	"os/exec"
)

// custom io.WriteCloser to handle on the fly mp3 convertion

const FFMPEG = "ffmpeg"

type cmdWriter struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func (w *cmdWriter) Write(p []byte) (n int, err error) {
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

// This class streams
func getFFplayWriter() (w *cmdWriter, err error) {
	_, err = exec.LookPath("ffplay")
	if err != nil {
		log.Fatalln("you need to install ffplay to run grank")
		os.Exit(1)
	}

	w = &cmdWriter{
		exec.Command("ffplay", "-nodisp", "-"),
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
	}
	w.stdin, err = w.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	w.cmd.Start()
	return w, nil
}

// This class will write the underlying stream (video) to a file
func getWriter(cfg *Config, stream stream) (out io.WriteCloser, err error) {
	path := cfg.OutputPath(stream)

	if _, err = os.Stat(path); err == nil && cfg.overwrite == false {
		return nil, fmt.Errorf("the destination file '%s' already exists and overwrite set to false", path)
	}

	if cfg.isMp3() {
		fmt.Printf("Converting video to mp3 file at '%s' ...\n", path)
		out, err = getFFmpegWriter(path, cfg.AudioBitrate(stream))
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
