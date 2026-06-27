package main

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"encoding/base64"
	"crypto/rand"


)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0o755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	key := make([]byte,32)
	rand.Read(key)
	keyString := base64.RawURLEncoding.EncodeToString(key)
	ext := mediaTypeToExt(mediaType)
	path := keyString + ext
	return fmt.Sprintf("%s",path)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}
