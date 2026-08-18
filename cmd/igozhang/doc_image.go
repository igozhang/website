package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func decodePasteImage(data string) ([]byte, string, error) {
	data = strings.TrimSpace(data)
	if !strings.HasPrefix(data, "data:") {
		return nil, "", fmt.Errorf("expected data URL")
	}
	i := strings.Index(data, ",")
	if i < 0 {
		return nil, "", fmt.Errorf("bad data URL")
	}
	hdr, b64 := data[:i], data[i+1:]
	ext := ""
	switch {
	case strings.Contains(hdr, "image/png"):
		ext = "png"
	case strings.Contains(hdr, "image/jpeg"), strings.Contains(hdr, "image/jpg"):
		ext = "jpg"
	default:
		return nil, "", fmt.Errorf("only png/jpeg/jpg allowed")
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, "", fmt.Errorf("invalid base64")
	}
	if len(raw) > docPasteImageMaxBytes {
		return nil, "", fmt.Errorf("image too large")
	}
	if !imageMagicOK(raw, ext) {
		return nil, "", fmt.Errorf("image content mismatch")
	}
	return raw, ext, nil
}

func imageMagicOK(raw []byte, ext string) bool {
	if len(raw) < 12 {
		return false
	}
	switch ext {
	case "png":
		return raw[0] == 0x89 && raw[1] == 0x50 && raw[2] == 0x4E && raw[3] == 0x47
	case "jpg":
		return raw[0] == 0xFF && raw[1] == 0xD8 && raw[2] == 0xFF
	}
	return false
}
