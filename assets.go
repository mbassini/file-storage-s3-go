package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	id := base64.RawURLEncoding.EncodeToString(base)

	fileExtension := mediaTypeToExtension(mediaType)
	return fmt.Sprintf("%s%s", id, fileExtension)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func (cfg apiConfig) getObjectURL(aspectRatio, assetPath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s/%s", cfg.s3Bucket, cfg.s3Region, aspectRatio, assetPath)
}

func mediaTypeToExtension(mediaType string) string {
	return "." + strings.Split(mediaType, "/")[1]
}

func (cfg apiConfig) getVideoAspectRatio(tempFile string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", tempFile)

	var output bytes.Buffer
	cmd.Stdout = &output

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error running ffprobe: %w", err)
	}

	type Stream struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	type ffprobeOutput struct {
		Streams []Stream `json:"streams"`
	}
	var ffOutput ffprobeOutput
	err = json.Unmarshal(output.Bytes(), &ffOutput)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling ffprobe output: %w", err)
	}
	if len(ffOutput.Streams) == 0 {
		return "", fmt.Errorf("No streams found")
	}

	firstStream := ffOutput.Streams[0]
	if firstStream.Height == 0 || firstStream.Width == 0 {
		return "", fmt.Errorf("Invalid stream dimensions: width=%d, height=%d", firstStream.Width, firstStream.Height)
	}

	w := float64(firstStream.Width)
	h := float64(firstStream.Height)
	actualRatio := w / h

	landscapeTargetRatio := 16.0 / 9.0
	portraitTargetRatio := 9.0 / 16.0

	diff := math.Abs(actualRatio-landscapeTargetRatio) / landscapeTargetRatio
	if diff < 0.05 {
		return "landscape", nil
	}

	diff = math.Abs(actualRatio-portraitTargetRatio) / portraitTargetRatio
	if diff < 0.05 {
		return "portrait", nil
	}

	return "other", nil
}
