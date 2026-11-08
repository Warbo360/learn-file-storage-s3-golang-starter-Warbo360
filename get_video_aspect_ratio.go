package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	// set up ffmprobe command to extract json info on the passed video file

	ffmprobe := exec.Command("ffmprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	// Set stout of the command to a new buffer

	var b bytes.Buffer
	ffmprobe.Stdout = &b

	// Run the command

	err := ffmprobe.Run()
	if err != nil {
		return "", errors.New("error running ffmprobe command")
	}

	// Define a struct to extract video width and height

	type AspectRatio struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	// Instantiate a newly defined aspectRatio struct, unpack the standard out into it

	var aspectRatio AspectRatio
	err = json.Unmarshal(b.Bytes(), &aspectRatio)
	if err != nil {
		return "", errors.New("error decoding stdout from ffmprobe")
	}

	// Calculate ratio and return based on that

	ratio := float64(aspectRatio.Width) / float64(aspectRatio.Height)
	if ratio >= 1.6 && ratio < 1.8 {
		return "16:9", nil
	} else if ratio >= 0.54 && ratio < 0.57 {
		return "9:16", nil
	}
	return "other", nil
}
