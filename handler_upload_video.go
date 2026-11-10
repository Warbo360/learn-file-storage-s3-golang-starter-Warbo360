package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	// Limit request Body read to 1GB

	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)

	// Obtain video ID from path value in the request

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	// Obtain JWT from request header

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not find JWT", err)
		return
	}

	// Validate JWT from previous step

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not validate JWT", err)
		return
	}

	// Obtain video metadata from database

	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not find video in database", err)
		return
	}

	// Validate video userID with JWT userID match

	if userID != dbVideo.UserID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	// Obtain file data and headers from request

	data, headers, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing file data and headers", err)
		return
	}
	defer data.Close()

	// Parse headers to get Content-Type header value to validate file is video/mp4

	headerMediaType := headers.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(headerMediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing header Content-Type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Incorrect file type in request", err)
		return
	}

	// Save uploaded video to a temp file on local disk

	file, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving temp upload", err)
		return
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := io.Copy(file, data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error copying video data to file", err)
		return
	}

	// Reset the file pointer to the start of the file to allow reading from the start again

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reset file pointer of saved video upload", err)
		return
	}

	// Generate key for s3 upload

	key := make([]byte, 32)
	rand.Read(key)
	encodedKey := base64.URLEncoding.EncodeToString(key) + ".mp4"

	// Get video aspect ratio and prefix file key with related prefix

	aspectRatio, err := getVideoAspectRatio(file.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing file for aspect ratio", err)
		return
	}
	fmt.Printf("aspectRatio: %v", aspectRatio)
	if aspectRatio == "16:9" {
		encodedKey = "/landscape/" + encodedKey
	} else if aspectRatio == "9:16" {
		encodedKey = "/portrait/" + encodedKey
	} else {
		encodedKey = "/other/" + encodedKey
	}

	// Actually uploading to s3

	input := s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &encodedKey,
		Body:        io.Reader(file),
		ContentType: &mediaType,
	}
	if _, err := cfg.s3Client.PutObject(
		context.Background(),
		&input,
	); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error putting file into s3", err)
		return
	}

	// Update database video URL

	updatedVideoURL := fmt.Sprintf("http://tubely-0864297531.s3.%v.amazonaws.com/%v", cfg.s3Region, encodedKey)
	dbVideo.VideoURL = &updatedVideoURL
	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating database video URL", err)
		return
	}
}
