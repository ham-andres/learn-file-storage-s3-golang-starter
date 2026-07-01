package main	

import (
		"bytes"
		"os/exec"
		"encoding/json"
		"math"
		"fmt"
		"os"
		
)

func getVideoAspectRatio(filePath string) (string, error) {
		
	type Stream struct {
			Width				int					`json:"width"`
			Height			int					`json:"height"`
	}
	type Response struct {
			Streams				[]Stream			`json:"streams"`
	}



		cmd := exec.Command("ffprobe",
												"-v", "error",
												"-print_format","json",
												"-show_streams",
												filePath,
											)
		var cmdOut bytes.Buffer
		cmd.Stdout = &cmdOut

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed while running the command: %v", err)
		}

		var probeResult Response
		err := json.Unmarshal(cmdOut.Bytes(), &probeResult)
		if err != nil {
			return "", fmt.Errorf("failed Unmarshaling: %v", err)
		}
		
		if len(probeResult.Streams) == 0 {
				return "", nil
		} 
		w := probeResult.Streams[0].Width
		h := probeResult.Streams[0].Height

		ratio := float64(w) / float64(h)
		
		
		tolerance := 0.01

				var aspectRatio string

		if math.Abs(ratio - (16.0/9.0)) < tolerance {
			aspectRatio = "16:9"
		} else if math.Abs(ratio - (9.0/16.0)) < tolerance {
			aspectRatio = "9:16"
		}	else {
				aspectRatio = "other"
		}

		return aspectRatio, nil 
}


func processVideoForFastStart( filePath string) (string, error) {
		var outFilePath string
		outFilePath = filePath + ".processing"

		cmd := exec.Command("ffmpeg",
												"-i", filePath,
												"-c","copy",
												"-movflags","faststart",
												"-f","mp4", outFilePath,
											)
		var cmdOut bytes.Buffer
		cmd.Stderr = &cmdOut

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("Error while cmd.Run(): %v", err)
		}

		info, err := os.Stat(outFilePath)
		if err != nil {
			return "", fmt.Errorf("file path might not exist in OS: %v",err)
		}
		if info.Size() == 0 {
			return "", fmt.Errorf("the file is empty: %v",err)
		}

		return outFilePath, nil 
	}
