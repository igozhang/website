package ipgeo

import (
	"encoding/binary"
	"fmt"
)

// Header igo-geo.ipgeo 文件头字段（大端，64 字节）。
type Header struct {
	Magic               [5]byte
	Version             uint8
	CountryOff          uint32
	CountryCount        uint16
	V4L1Off, V4L1Count uint32
	V4L2Off, V4L2Count uint32
	V6WideOff, V6WideCount       uint32
	V6CompactOff, V6CompactCount uint32
}

// ParseHeader 解析文件头；不校验文件整体有效性。
func ParseHeader(data []byte) (Header, error) {
	if len(data) < headerSize {
		return Header{}, fmt.Errorf("ipgeo: file too small")
	}
	return Header{
		Magic:          [5]byte{data[0], data[1], data[2], data[3], data[4]},
		Version:        data[5],
		CountryOff:     binary.BigEndian.Uint32(data[8:12]),
		CountryCount:   binary.BigEndian.Uint16(data[12:14]),
		V4L1Off:        binary.BigEndian.Uint32(data[16:20]),
		V4L1Count:      binary.BigEndian.Uint32(data[20:24]),
		V4L2Off:        binary.BigEndian.Uint32(data[24:28]),
		V4L2Count:      binary.BigEndian.Uint32(data[28:32]),
		V6WideOff:      binary.BigEndian.Uint32(data[32:36]),
		V6WideCount:    binary.BigEndian.Uint32(data[36:40]),
		V6CompactOff:   binary.BigEndian.Uint32(data[40:44]),
		V6CompactCount: binary.BigEndian.Uint32(data[44:48]),
	}, nil
}

// V4RecordRaw 从二进制偏移读取 IPv4 索引条（8B）。
func V4RecordRaw(data []byte, off int) (start uint32, prefix uint8, ccIdx uint16, pad uint8) {
	p := off
	start = binary.BigEndian.Uint32(data[p : p+4])
	prefix = data[p+4]
	ccIdx = binary.BigEndian.Uint16(data[p+5 : p+7])
	pad = data[p+7]
	return
}

// V6WideRecordRaw 从二进制偏移读取 IPv6 wide 条（20B）。
func V6WideRecordRaw(data []byte, off int) (start [16]byte, prefix uint8, ccIdx uint16, pad uint8) {
	copy(start[:], data[off:off+16])
	prefix = data[off+16]
	ccIdx = binary.BigEndian.Uint16(data[off+17 : off+19])
	pad = data[off+19]
	return
}

// V6CompactRecordRaw 从二进制偏移读取 IPv6 compact 条（10B）。
func V6CompactRecordRaw(data []byte, off int) (prefix uint8, start [6]byte, ccIdx uint16, pad uint8) {
	prefix = data[off]
	copy(start[:], data[off+1:off+7])
	ccIdx = binary.BigEndian.Uint16(data[off+7 : off+9])
	pad = data[off+9]
	return
}

// CountryRaw 从国家表偏移解析单条（返回下一偏移）。
func CountryRaw(data []byte, off int) (c Country, flags uint8, next int, err error) {
	if off+6 > len(data) {
		return Country{}, 0, off, fmt.Errorf("truncated country header")
	}
	cc := string(data[off : off+2])
	off += 2
	cont := string(data[off : off+2])
	off += 2
	flags = data[off]
	off++
	zl := int(data[off])
	off++
	if off+zl > len(data) {
		return Country{}, flags, off, fmt.Errorf("zh name out of range")
	}
	zn := string(data[off : off+zl])
	off += zl
	if off >= len(data) {
		return Country{}, flags, off, fmt.Errorf("missing en len")
	}
	el := int(data[off])
	off++
	if off+el > len(data) {
		return Country{}, flags, off, fmt.Errorf("en name out of range")
	}
	en := string(data[off : off+el])
	off += el
	return Country{
		Code:           cc,
		ContinentCode:  cont,
		NameZH:         zn,
		NameEN:         en,
		ChinesePrimary: flags&1 != 0,
	}, flags, off, nil
}
