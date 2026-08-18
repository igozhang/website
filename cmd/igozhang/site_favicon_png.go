package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"sync"
)

// siteFaviconPNG：锁屏/通知栏 Media Session 用位图（Android Chrome 对 SVG artwork 支持差）。
var (
	faviconPNGOnce  sync.Once
	faviconPNGBytes []byte
)

func siteFaviconPNG() []byte {
	faviconPNGOnce.Do(func() {
		const S = 512
		img := image.NewRGBA(image.Rect(0, 0, S, S))
		bg := color.RGBA{0x1a, 0x2d, 0x1a, 0xff}
		for y := 0; y < S; y++ {
			for x := 0; x < S; x++ {
				img.Set(x, y, bg)
			}
		}
		scale := S / 16
		fillRect := func(x0, y0, w, h int, c color.Color) {
			x1, y1 := x0*scale, y0*scale
			x2, y2 := (x0+w)*scale, (y0+h)*scale
			for y := y1; y < y2 && y < S; y++ {
				for x := x1; x < x2 && x < S; x++ {
					if x >= 0 && y >= 0 {
						img.Set(x, y, c)
					}
				}
			}
		}
		fillRect(5, 2, 6, 2, color.RGBA{0xc4, 0x5c, 0x3a, 0xff})
		fillRect(4, 4, 8, 3, color.RGBA{0x4a, 0x7a, 0x4a, 0xff})
		fillRect(5, 7, 2, 4, color.RGBA{0x3d, 0x5c, 0x8a, 0xff})
		fillRect(9, 7, 2, 4, color.RGBA{0x3d, 0x5c, 0x8a, 0xff})
		fillRect(6, 7, 4, 2, color.RGBA{0x5a, 0x8a, 0x5a, 0xff})
		fillRect(6, 9, 4, 3, color.RGBA{0x2d, 0x4a, 0x2d, 0xff})
		fillRect(3, 5, 2, 1, color.RGBA{0x8a, 0x5a, 0x3a, 0xff})
		fillRect(11, 5, 3, 1, color.RGBA{0x5a, 0x5a, 0x5a, 0xff})
		fillRect(12, 6, 2, 1, color.RGBA{0x3a, 0x3a, 0x3a, 0xff})
		fillRect(6, 12, 1, 2, color.RGBA{0x2a, 0x3d, 0x5a, 0xff})
		fillRect(9, 12, 1, 2, color.RGBA{0x2a, 0x3d, 0x5a, 0xff})
		fillRect(5, 14, 2, 1, color.RGBA{0x1a, 0x1a, 0x1a, 0xff})
		fillRect(9, 14, 2, 1, color.RGBA{0x1a, 0x1a, 0x1a, 0xff})
		var buf bytes.Buffer
		_ = png.Encode(&buf, img)
		faviconPNGBytes = buf.Bytes()
	})
	return faviconPNGBytes
}

func writeSiteFaviconPNG(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(siteFaviconPNG())
}
