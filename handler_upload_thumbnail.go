package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

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

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()
	ctHeader := header.Header.Get("Content-Type")
	ct, _, err := mime.ParseMediaType(ctHeader)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to identificar content type header",
			err,
		)
		return
	}

	if !validContentTypes(ct) {
		respondWithError(w, http.StatusBadRequest, "content type invalido", nil)
		return
	}

	fileExt := getFileExtension(ct)
	key := make([]byte, 32)
	rand.Read(key)
	// fileName:=base64.RawURLEncoding.EncodeToString(key)
	videoFileName := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(key), fileExt)
	fullVideoPathName := filepath.Join(cfg.assetsRoot, videoFileName)
	videoFile, err := os.Create(fullVideoPathName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to crear archivo de video", err)
		return
	}
	io.Copy(videoFile, file)
	// thumbBytes, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Unable to read form file", err)
	// 	return
	// }

	// thumbURL := fmt.Sprintf(
	// 	"data:%s;base64,%s",
	// 	contentType,
	// 	base64.StdEncoding.EncodeToString(thumbBytes),
	// )

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

	// videoThumbnails[videoID] = thumbnail{data: thumbBytes, mediaType: contentType}
	// thumbUrl := fmt.Sprintf("http://10.10.111.3:8091/api/thumbnails/%s", videoID.String())
	thumbURL := fmt.Sprintf(
		"http://10.10.111.3:%s/assets/%s",
		cfg.port,
		// videoID.String(),
		videoFileName,
	)
	videoMetadata.ThumbnailURL = &thumbURL
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

func validContentTypes(ct string) bool {
	vct := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}
	valid, ok := vct[ct]
	if ok {
		return valid
	}
	return false
}

func getFileExtension(ct string) string {
	dicExtension := map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
		"video/mp4":  "mp4",
	}
	ext, ok := dicExtension[ct]
	if ok {
		return ext
	}
	return "bin"
}
