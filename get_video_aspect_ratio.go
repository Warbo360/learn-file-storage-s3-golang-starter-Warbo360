package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
)

// Define a struct to extract video width and height from stdout stream

type Stdout struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	// set up ffmprobe command to extract json info on the passed video file

	ffprobe := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	// Set stout of the command to a new buffer

	var b bytes.Buffer
	ffprobe.Stdout = &b

	// Run the command

	err := ffprobe.Run()
	if err != nil {
		return "", errors.New("error running ffprobe command")
	}

	// Instantiate a newly defined aspectRatio struct, unpack the standard out into it

	var stdout Stdout
	err = json.Unmarshal(b.Bytes(), &stdout)
	if err != nil {
		return "", errors.New("error decoding stdout from ffmprobe")
	}

	// Calculate ratio and return based on that

	ratio := float64(stdout.Streams[0].Width) / float64(stdout.Streams[0].Height)
	if ratio >= 1.6 && ratio < 1.8 {
		return "16:9", nil
	} else if ratio >= 0.54 && ratio < 0.57 {
		return "9:16", nil
	}
	return "other", nil
}
