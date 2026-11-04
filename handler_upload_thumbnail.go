package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	// imageSlice, err := io.ReadAll(data)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Error reading image file", err)
	// 	return
	// }

	// NOTE: --- HERE DOWN IS FOR UPLOADING THUMBNAILS VIA BASE64 STRING ENCODING INTO SQLITE

	// Encode imageSLice to base64 string to store in relational database

	// imageString64 := base64.StdEncoding.EncodeToString(imageSlice)

	// Create dataURL string

	// imageDataURL := fmt.Sprintf("data:%v;base64,%v", headerMediaType, imageString64)

	// NOTE: --- HERE UP IS FOR UPLOADING THUMBNAILS VIA BASE64 STRING ENCODING INTO SQLITE

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

	// Extract file extension needed for path from content-type

	extension, err := extractFileExtension(headerMediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse content type", err)
		return
	}

	// Create filepath to save thumbnail image to disk

	imagePath := filepath.Join(cfg.assetsRoot, dbVideo.ID.String()+"."+extension)

	// Create the new file to save to disk

	file, err := os.Create(imagePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating filepath to save image to", err)
		return
	}
	if _, err := io.Copy(file, data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving image bytes to file", err)
		return
	}

	// NOTE: --- HERE IS FOR UPLOADINF THUMBNAILS VIA GLOBAL MAP

	// Create new thumbnail struct with obtained image slice and media type

	// newTN := thumbnail{
	// 	data:      imageSlice,
	// 	mediaType: headerMediaType,
	// }

	// Add the new thumbnail to the global map

	//videoThumbnails[dbVideo.ID] = newTN

	// Construct the new thumbnail url and set that for the dbVideo

	//newURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, dbVideo.ID.String())

	// NOTE: --- UP IS FOR UPLOADING THUMBNAILS VIA GLOBAL MAP

	tnURL := fmt.Sprintf("http://localhost:%v/%v", cfg.port, imagePath)
	dbVideo.ThumbnailURL = &tnURL

	// Query to update the video after updating thumbnail URL

	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save updated thumbnail", err)
		return
	}

	// Respond with updated dbVideo struct

	respondWithJSON(w, http.StatusOK, dbVideo)
}

func extractFileExtension(s string) (string, error) {
	splitContentType := strings.Split(s, "/")
	if len(splitContentType) != 2 {
		return "", errors.New("Content-Type not supported by extractFileExtension")
	}
	return splitContentType[1], nil
}
