package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	// set max memory and parse multipart form on the response

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse mutlipart form", err)
		return
	}

	// Obtain file data and headers from "thumbnail key"

	data, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not retrieve file and header", err)
		return
	}
	defer data.Close()

	// Obtain content-type header value

	headerMediaType := header.Header.Get("Content-Type")

	// Read image data into a bute slice

	imageSlice, err := io.ReadAll(data)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error reading image file", err)
		return
	}

	// Encode imageSLice to base64 string to store in relational database

	imageString64 := base64.StdEncoding.EncodeToString(imageSlice)

	// Create dataURL string

	imageDataURL := fmt.Sprintf("data:%v;base64,%v", headerMediaType, imageString64)

	// Get db hit on video ID

	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video not found", err)
		return
	}

	// Check User ID for db hit on video matches what was returned for JWT validate

	if dbVideo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		return
	}

	// Create new thumbnail struct with obtained image slice and media type

	// newTN := thumbnail{
	// 	data:      imageSlice,
	// 	mediaType: headerMediaType,
	// }

	// Add the new thumbnail to the global map

	//videoThumbnails[dbVideo.ID] = newTN

	// Construct the new thumbnail url and set that for the dbVideo

	//newURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, dbVideo.ID.String())
	dbVideo.ThumbnailURL = &imageDataURL

	// Query to update the video after updating thumbnail URL

	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save updated thumbnail", err)
		return
	}

	// Respond with updated dbVideo struct

	respondWithJSON(w, http.StatusOK, dbVideo)
}
