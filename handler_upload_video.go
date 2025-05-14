package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
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

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "usuario no autorizado", err)
		return
	}

	fmt.Println("uploading Video", videoID, "by user", userID)

	const maxFileSize = 1 << 30 // 1G
	const maxMemory = 10 << 20  // 10M
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file 1", err)
		return
	}
	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file 2", err)
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

	if !validVideoContentTypes(ct) {
		respondWithError(w, http.StatusBadRequest, "content type invalido", nil)
		return
	}

	// fileExt := getFileExtension(ct)
	// fileName:=base64.RawURLEncoding.EncodeToString(key)
	// videoFileName := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(key), fileExt)
	// fullVideoPathName := filepath.Join(cfg.assetsRoot, videoFileName)
	// videoFile, err := os.Create(fullVideoPathName)
	videoFile, err := os.CreateTemp("", "tubely.upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to crear archivo de video", err)
		return
	}
	defer os.Remove(videoFile.Name())
	defer videoFile.Close()

	io.Copy(videoFile, file)
	videoFile.Seek(0, io.SeekStart)

	fileExt := getFileExtension(ct)
	key := make([]byte, 32)
	rand.Read(key)
	// aspect, err := GetVideoAspectRatio(fmt.Sprintf("%s/tubely.upload.mp4", os.TempDir()))
	aspect, err := GetVideoAspectRatio(videoFile.Name())
	if err != nil {
		slog.Warn("Get Aspect", "error", err)
	}

	aspPrefix := "other"
	if aspect == "16:9" {
		aspPrefix = "landscape"
	}

	if aspect == "9:16" {
		aspPrefix = "portrait"
	}
	videoFileName := fmt.Sprintf(
		"%s/%s.%s",
		aspPrefix,
		base64.RawURLEncoding.EncodeToString(key),
		fileExt,
	)
	objectParams := s3.PutObjectInput{
		Key:         &videoFileName,
		Bucket:      &cfg.s3Bucket,
		ContentType: &ct,
		Body:        videoFile,
	}
	cfg.s3Client.PutObject(r.Context(), &objectParams)

	videoURL := fmt.Sprintf(
		"https://%s.s3.%s.amazonaws.com/%s",
		cfg.s3Bucket,
		cfg.s3Region,
		videoFileName,
	)
	videoMetadata.VideoURL = &videoURL

	err = cfg.db.UpdateVideo(videoMetadata)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to crear archivo de video en S3",
			err,
		)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMetadata)
}

func validVideoContentTypes(ct string) bool {
	vct := map[string]bool{
		"video/mp4": true,
	}
	valid, ok := vct[ct]
	if ok {
		return valid
	}
	return false
}
