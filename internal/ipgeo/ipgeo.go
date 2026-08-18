package ipgeo

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

const headerSize = 64

// Country 国家元数据。
type Country struct {
	Code, ContinentCode string
	NameZH, NameEN      string
	ChinesePrimary      bool
}

// Name 返回国家显示名；lang 为 zh、en 或 auto（默认）。
func (c Country) Name(lang string) string {
	switch lang {
	case "zh":
		if c.NameZH != "" {
			return c.NameZH
		}
		return c.NameEN
	case "en":
		if c.NameEN != "" {
			return c.NameEN
		}
		return c.NameZH
	default:
		if c.ChinesePrimary {
			return c.NameZH
		}
		return c.NameEN
	}
}

// Result 查询结果。
type Result struct {
	Country Country
	CIDR    string
}

// DB IP 地理库。
type DB struct {
	countries []Country
	v4L1      []v4Rec
	v4L2      []v4Rec
	v6Wide    []v6WideRec
	v6Compact []v6CompactRec
}

type v4Rec struct {
	start uint32
	pfx   uint8
	cc    uint16
}

type v6WideRec struct {
	start [16]byte
	pfx   uint8
	cc    uint16
}

type v6CompactRec struct {
	key [6]byte
	pfx uint8
	cc  uint16
}

// Open 从磁盘加载 ipgeo 二进制。
func Open(path string) (*DB, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return OpenBytes(data)
}

// OpenBytes 从内存加载 ipgeo 二进制。
func OpenBytes(data []byte) (*DB, error) {
	if len(data) < headerSize || string(data[:5]) != "IPGEO" || data[5] != 2 {
		return nil, fmt.Errorf("ipgeo: invalid file")
	}
	u32 := func(o int) uint32 { return binary.BigEndian.Uint32(data[o : o+4]) }
	u16 := func(o int) uint16 { return binary.BigEndian.Uint16(data[o : o+2]) }

	countryOff := int(u32(8))
	countryN := int(u16(12))
	v4L1Off, v4L1N := int(u32(16)), int(u32(20))
	v4L2Off, v4L2N := int(u32(24)), int(u32(28))
	v6WideOff, v6WideN := int(u32(32)), int(u32(36))
	v6CompactOff, v6CompactN := int(u32(40)), int(u32(44))

	countries, err := parseCountries(data, countryOff, countryN, v4L1Off)
	if err != nil {
		return nil, err
	}
	if err := boundsCheck(data, v4L1Off, v4L1N, 8); err != nil {
		return nil, fmt.Errorf("ipgeo: v4 L1: %w", err)
	}
	if err := boundsCheck(data, v4L2Off, v4L2N, 8); err != nil {
		return nil, fmt.Errorf("ipgeo: v4 L2: %w", err)
	}
	if err := boundsCheck(data, v6WideOff, v6WideN, 20); err != nil {
		return nil, fmt.Errorf("ipgeo: v6 wide: %w", err)
	}
	if err := boundsCheck(data, v6CompactOff, v6CompactN, 10); err != nil {
		return nil, fmt.Errorf("ipgeo: v6 compact: %w", err)
	}

	return &DB{
		countries: countries,
		v4L1:      parseV4(data, v4L1Off, v4L1N),
		v4L2:      parseV4(data, v4L2Off, v4L2N),
		v6Wide:    parseV6Wide(data, v6WideOff, v6WideN),
		v6Compact: parseV6Compact(data, v6CompactOff, v6CompactN),
	}, nil
}

func boundsCheck(data []byte, off, n, recSize int) error {
	if off < 0 || n < 0 || recSize <= 0 {
		return fmt.Errorf("bad offset/count")
	}
	end := off + n*recSize
	if end > len(data) {
		return fmt.Errorf("segment out of range")
	}
	return nil
}

func parseCountries(data []byte, off, count, endAt int) ([]Country, error) {
	out := make([]Country, 0, count)
	start := off
	for i := 0; i < count; i++ {
		if off+6 > len(data) {
			return nil, fmt.Errorf("country %d: truncated header", i)
		}
		cc := string(data[off : off+2])
		off += 2
		cont := string(data[off : off+2])
		off += 2
		flags := data[off]
		off++
		zl := int(data[off])
		off++
		if off+zl > len(data) {
			return nil, fmt.Errorf("country %d %s: zh name out of range", i, cc)
		}
		zn := string(data[off : off+zl])
		off += zl
		if off >= len(data) {
			return nil, fmt.Errorf("country %d %s: missing en len", i, cc)
		}
		el := int(data[off])
		off++
		if off+el > len(data) {
			return nil, fmt.Errorf("country %d %s: en name out of range", i, cc)
		}
		en := string(data[off : off+el])
		off += el
		out = append(out, Country{
			Code:           cc,
			ContinentCode:  cont,
			NameZH:         zn,
			NameEN:         en,
			ChinesePrimary: flags&1 != 0,
		})
	}
	if off != endAt {
		return nil, fmt.Errorf("country table size mismatch: parsed to %d want %d", off, endAt)
	}
	if count > 0 && start == off {
		return nil, fmt.Errorf("empty country table")
	}
	return out, nil
}

func parseV4(data []byte, off, n int) []v4Rec {
	out := make([]v4Rec, n)
	for i := 0; i < n; i++ {
		p := off + i*8
		out[i] = v4Rec{
			start: binary.BigEndian.Uint32(data[p : p+4]),
			pfx:   data[p+4],
			cc:    binary.BigEndian.Uint16(data[p+5 : p+7]),
		}
	}
	return out
}

func parseV6Wide(data []byte, off, n int) []v6WideRec {
	out := make([]v6WideRec, n)
	for i := 0; i < n; i++ {
		p := off + i*20
		var start [16]byte
		copy(start[:], data[p:p+16])
		out[i] = v6WideRec{
			start: start,
			pfx:   data[p+16],
			cc:    binary.BigEndian.Uint16(data[p+17 : p+19]),
		}
	}
	return out
}

func parseV6Compact(data []byte, off, n int) []v6CompactRec {
	out := make([]v6CompactRec, n)
	for i := 0; i < n; i++ {
		p := off + i*10
		var key [6]byte
		copy(key[:], data[p+1:p+7])
		out[i] = v6CompactRec{
			pfx: data[p],
			key: key,
			cc:  binary.BigEndian.Uint16(data[p+7 : p+9]),
		}
	}
	return out
}

// Lookup 查询 IP 所属国家与大洲。
func (db *DB) Lookup(ipStr string) (*Result, bool) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, false
	}
	if v4 := ip.To4(); v4 != nil {
		n := binary.BigEndian.Uint32(v4)
		if r, ok := searchV4(db.v4L1, db.countries, n); ok {
			return r, true
		}
		return searchV4(db.v4L2, db.countries, n)
	}
	return db.lookupV6(ip)
}

func searchV4(recs []v4Rec, countries []Country, n uint32) (*Result, bool) {
	if len(recs) == 0 {
		return nil, false
	}
	lo, hi := 0, len(recs)-1
	best := -1
	for lo <= hi {
		mid := (lo + hi) / 2
		if recs[mid].start <= n {
			best = mid
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	if best < 0 {
		return nil, false
	}
	r := recs[best]
	if !v4Match(n, r.start, r.pfx) {
		return nil, false
	}
	return resultFromCC(countries, r.cc, r.start, r.pfx, false)
}

func (db *DB) lookupV6(ip net.IP) (*Result, bool) {
	ip16 := ip.To16()
	if ip16 == nil {
		return nil, false
	}
	var best *Result
	var bestPfx uint8
	if r, pfx, ok := searchV6Wide(db.v6Wide, db.countries, ip16); ok && pfx > bestPfx {
		best, bestPfx = r, pfx
	}
	if r, pfx, ok := searchV6Compact(db.v6Compact, db.countries, ip16); ok && pfx > bestPfx {
		best, bestPfx = r, pfx
	}
	if best != nil {
		return best, true
	}
	return nil, false
}

func searchV6Wide(recs []v6WideRec, countries []Country, ip16 net.IP) (*Result, uint8, bool) {
	if len(recs) == 0 {
		return nil, 0, false
	}
	key := ip16ToKey(ip16)
	lo, hi := 0, len(recs)-1
	best := -1
	for lo <= hi {
		mid := (lo + hi) / 2
		if bytesLE(recs[mid].start[:], key[:]) {
			best = mid
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	if best < 0 {
		return nil, 0, false
	}
	r := recs[best]
	if !v6Match(ip16, r.start[:], r.pfx) {
		return nil, 0, false
	}
	res, ok := resultFromCC(countries, r.cc, 0, r.pfx, true)
	if !ok {
		return nil, 0, false
	}
	res.CIDR = v6CIDR(r.start[:], r.pfx)
	return res, r.pfx, true
}

func searchV6Compact(recs []v6CompactRec, countries []Country, ip16 net.IP) (*Result, uint8, bool) {
	if len(recs) == 0 {
		return nil, 0, false
	}
	var key [16]byte
	copy(key[:], ip16)
	lo, hi := 0, len(recs)-1
	best := -1
	for lo <= hi {
		mid := (lo + hi) / 2
		if bytesLE(recs[mid].key[:], key[:6]) {
			best = mid
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	if best < 0 {
		return nil, 0, false
	}
	r := recs[best]
	var start [16]byte
	copy(start[:6], r.key[:])
	if !v6Match(ip16, start[:], r.pfx) {
		return nil, 0, false
	}
	res, ok := resultFromCC(countries, r.cc, 0, r.pfx, true)
	if !ok {
		return nil, 0, false
	}
	res.CIDR = v6CIDR(start[:], r.pfx)
	return res, r.pfx, true
}

func resultFromCC(countries []Country, ccIdx uint16, start uint32, pfx uint8, v6 bool) (*Result, bool) {
	if int(ccIdx) >= len(countries) {
		return nil, false
	}
	c := countries[ccIdx]
	var cidr string
	if v6 {
		cidr = ""
	} else {
		cidr = v4CIDR(start, pfx)
	}
	return &Result{Country: c, CIDR: cidr}, true
}

func v4Match(ip, start uint32, prefix uint8) bool {
	if prefix == 0 {
		return true
	}
	if prefix > 32 {
		return false
	}
	mask := uint32(0)
	if prefix == 32 {
		mask = ^uint32(0)
	} else {
		mask = ^uint32(0) << (32 - prefix)
	}
	return (ip & mask) == (start & mask)
}

func v6Match(ip net.IP, start []byte, prefix uint8) bool {
	if prefix == 0 {
		return true
	}
	if prefix > 128 {
		return false
	}
	ip16 := ip.To16()
	if ip16 == nil || len(start) < 16 {
		return false
	}
	full := int(prefix) / 8
	rem := int(prefix) % 8
	if full > 16 {
		full = 16
	}
	if full > 0 && !bytesEqual(ip16[:full], start[:full]) {
		return false
	}
	if rem == 0 || full >= 16 {
		return true
	}
	mask := byte(0xFF << (8 - rem))
	return (ip16[full] & mask) == (start[full] & mask)
}

func v4CIDR(start uint32, prefix uint8) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, start)
	if prefix > 32 {
		prefix = 32
	}
	return fmt.Sprintf("%s/%d", ip.String(), prefix)
}

func v6CIDR(start []byte, prefix uint8) string {
	ip := make(net.IP, 16)
	copy(ip, start)
	if prefix > 128 {
		prefix = 128
	}
	return fmt.Sprintf("%s/%d", ip.String(), prefix)
}

func ip16ToKey(ip net.IP) [16]byte {
	var k [16]byte
	copy(k[:], ip.To16())
	return k
}

func bytesLE(a, b []byte) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return len(a) <= len(b)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Stats 返回库内各段条数（用于校验）。
func (db *DB) Stats() (countries, v4L1, v4L2, v6Wide, v6Compact int) {
	return len(db.countries), len(db.v4L1), len(db.v4L2), len(db.v6Wide), len(db.v6Compact)
}

// MaxCountryIndex 返回索引中引用的最大国家下标。
func (db *DB) MaxCountryIndex() uint16 {
	var max uint16
	check := func(idx uint16) {
		if idx > max {
			max = idx
		}
	}
	for _, r := range db.v4L1 {
		check(r.cc)
	}
	for _, r := range db.v4L2 {
		check(r.cc)
	}
	for _, r := range db.v6Wide {
		check(r.cc)
	}
	for _, r := range db.v6Compact {
		check(r.cc)
	}
	return max
}

// CountryAt 按索引取国家（测试用）。
func (db *DB) CountryAt(i int) Country {
	return db.countries[i]
}

// CountryByCode 按 ISO 3166-1 alpha-2 代码取国家元数据。
func (db *DB) CountryByCode(code string) (Country, bool) {
	if db == nil {
		return Country{}, false
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return Country{}, false
	}
	for _, c := range db.countries {
		if c.Code == code {
			return c, true
		}
	}
	return Country{}, false
}
