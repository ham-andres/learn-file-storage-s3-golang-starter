package main

import (
	"fmt"
	"net/http"
	"io"
	"os"
	"mime"


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
	const maxMemory = 10 << 20 // 10 MB
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
			respondWithError(w, http.StatusBadRequest, "couldn't get thumbnail", err)
			return
	}
	defer file.Close()
	
	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
			respondWithError(w, http.StatusBadRequest, "Couldn't access Content-Type", err)
			return
	}
	if mediaType != "image/png" && mediaType != "image/jpeg" {
			respondWithError(w, http.StatusBadRequest, "Invalid file type only images: png or jpeg", err)
			return
	}
	// we are doing this part for verification of the user and the related video
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Video not in database", err)
			return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}
	
	assetPath := getAssetPath(mediaType)	
	filePath := cfg.getAssetDiskPath(assetPath)
// maybe we can do this directly. create file first.
	
	fileDir, err := os.Create(filePath)
	if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create destination folder in local", err)
			return
	}
	defer fileDir.Close()

	_, err = io.Copy(fileDir, file )
	if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't save the thumbnail", err)
			return
	}
	
	url := cfg.getAssetURL(assetPath) 
	video.ThumbnailURL = &url

	err = cfg.db.UpdateVideo(video)
	if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't update the video", err)
			return
	}




	respondWithJSON(w, http.StatusOK, video)
}
