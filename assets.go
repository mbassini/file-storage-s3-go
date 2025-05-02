package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(videoID uuid.UUID, mediaType string) string {
	fileExtension := mediaTypeToExtension(mediaType)
	return fmt.Sprintf("%s%s", videoID, fileExtension)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExtension(mediaType string) string {
	return "." + strings.Split(mediaType, "/")[1]
}
