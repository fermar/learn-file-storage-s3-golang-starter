package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
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

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()
	contentType := header.Header.Get("Content-Type")
	thumbBytes, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read form file", err)
		return
	}
	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to get video metadata from db",
			err,
		)
	}
	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Usuario no autorizado", err)
		return
	}
	videoThumbnails[videoID] = thumbnail{data: thumbBytes, mediaType: contentType}
	thumbUrl := fmt.Sprintf("http://10.10.111.3:8091/api/thumbnails/%s", videoID.String())
	videoMetadata.ThumbnailURL = &thumbUrl
	err = cfg.db.UpdateVideo(videoMetadata)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to Update video metadata to db",
			err,
		)
	}
	// respondWithJSON(w, http.StatusOK, struct{}{})
	respondWithJSON(w, http.StatusOK, videoMetadata)
}
