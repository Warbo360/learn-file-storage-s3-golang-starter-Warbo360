package main

import (
	"mime"
	"net/http"

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
}
