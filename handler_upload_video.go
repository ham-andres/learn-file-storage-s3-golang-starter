package main

import (
	"net/http"
	"mime"
	"os"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
		const uploadLimit = 1 << 30 // max limit 1GB
		r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)
		
		videoIDString := r.PathValue("videoID")
		videoID, err := uuid.Parse(videoIDString)
		if err != nil {
				respondWithError(w, http.StatusBadRequest, "invalid ID request", err)
				return
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Couldn't fin JWT", err)
				return
		}

		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
		if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
				return
		}

		// authentication
		video, err := cfg.db.GetVideo(videoID)
		if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Video not in database", err)
				return
		}

		if video.UserID != userID {
				respondWithError(w, http.StatusUnauthorized, "unauthorized video access denied", nil)
				return
		}

		file, header , err := r.FormFile("video")
		if err != nil {
				respondWithError(w, http.StatusBadRequest, "Couldn't get video", err)
				return
		}
		defer file.Close()

		mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
		if err != nil {
				respondWithError(w, http.StatusBadRequest, "Couldn't access Content-Type", err)
				return
		}

		if mediaType != "video/mp4" {
				respondWithError(w, http.StatusBadRequest, "Invalid file type, only mp4", err)
				return
		}
		
		// "" in createtemp is for default tmp directory
		tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
		if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't find destination directory", err)
				return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		/* why are we doing io.Copy ?
			file ia the uploaded video coming from the http request, and it lives on memory/network stream 
			io.Copy writes it to tempFile on disk. we need on disk for processing ffmpeg
		*/
		if _, err = io.Copy(tempFile, file); err != nil {
				respondWithError(w, http.StatusInternalServerError, "storing to tempFile failed", err)
				return
		}
		// this below part is required for pointer to go back to start after io.Copy moves it to end 
		// else other operation would read an blank of the end.
		if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't reset file pointer", err)
				return
		}
		
		aspectRatio, err := getVideoAspectRatio(tempFile.Name())
		if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Could not get video aspect ratio", err)
				return
		}

		var videoFormat string

		if aspectRatio == "16:9" {
				videoFormat = "landscape"
		} else if aspectRatio == "9:16" {
				videoFormat = "portrait"
		} else {
				videoFormat = "other"
		}

		processTempFile, err := processVideoForFastStart(tempFile.Name())
		if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Could not get the processed temp file", err)
				return
		}

		videoForStreaming, err := os.Open(processTempFile); 
		if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Could not process video for fast streaming", err)
				return
		}

		defer os.Remove(processTempFile)
		defer videoForStreaming.Close()

		//Key
		assetPath := path.Join(videoFormat, getAssetPath(mediaType))


		if _, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
				Bucket:					aws.String(cfg.s3Bucket),
				Key:						aws.String(assetPath),
				Body:						videoForStreaming,
				ContentType: 		aws.String(mediaType),
		}); err != nil {
				respondWithError(w, http.StatusInternalServerError, "Could not get object passed in aws s3", err)
				return
		}

		url := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", cfg.s3Bucket, cfg.s3Region, assetPath)
		video.VideoURL = &url

		err = cfg.db.UpdateVideo(video)
		if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Could not update the video", err)
				return
		}

		respondWithJSON(w, http.StatusOK, video)

}
