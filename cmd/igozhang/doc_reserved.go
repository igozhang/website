package main

import "strings"

// docReservedSlug blocks article URLs that collide with site routes or assets.
func docReservedSlug(seg string) bool {
	seg = strings.ToLower(strings.TrimSpace(seg))
	if seg == "" {
		return true
	}
	_, ok := docReservedSlugs[seg]
	return ok
}

var docReservedSlugs = map[string]bool{
	"ops": true, "fs": true, "api": true, "health": true,
	"files": true, "preview": true, "stream": true,
	"gomoku": true, "xiangqi": true, "texas": true, "zhijian": true, "igokey": true, "igoadmin": true, "mermaid": true,
	"assets": true, "banner": true, "banner.png": true, "banner.jpg": true, "banner.jpeg": true,
	"favicon": true, "favicon.svg": true, "favicon.png": true, "favicon.jpg": true, "favicon.jpeg": true,
	"texas-card-back": true, "texas-card-back.png": true, "texas-card-back.jpg": true, "texas-card-back.jpeg": true,
	"mp3": true, "md": true, "doc-pic": true,
}
