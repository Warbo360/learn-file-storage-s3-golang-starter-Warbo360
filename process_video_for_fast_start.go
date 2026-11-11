package main

import (
	"errors"
	"os/exec"
)

func processVideoForFastStart(filepath string) (string, error) {
	// Output filepath
	output := filepath + ".processing"

	// Build command for ffmpeg

	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", output)
	err := cmd.Run()
	if err != nil {
		return "", errors.New("error running ffmpeg command")
	}
	return output, nil
}
