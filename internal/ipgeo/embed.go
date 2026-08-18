package ipgeo

import _ "embed"

//go:embed igo-geo-embed.ipgeo
var embedded []byte

// OpenEmbedded 从内置 igo-geo-embed.ipgeo 加载地理库。
func OpenEmbedded() (*DB, error) {
	return OpenBytes(embedded)
}
