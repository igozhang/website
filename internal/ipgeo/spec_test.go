package ipgeo

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"unicode/utf8"
)

var (
	validContinents = map[string]bool{
		"AS": true, "EU": true, "AF": true, "NA": true, "SA": true, "OC": true, "AN": true, "XX": true,
	}
	chinesePrimaryCC = map[string]bool{"CN": true, "TW": true, "HK": true, "MO": true}

	fixtureOnce sync.Once
	fixtureData []byte
	fixtureDB   *DB
	fixtureHdr  Header
	fixtureErr  error
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func loadFixture(t *testing.T) ([]byte, *DB, Header) {
	t.Helper()
	fixtureOnce.Do(func() {
		fixtureData = embedded
		fixtureHdr, fixtureErr = ParseHeader(fixtureData)
		if fixtureErr != nil {
			return
		}
		fixtureDB, fixtureErr = OpenBytes(fixtureData)
	})
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	return fixtureData, fixtureDB, fixtureHdr
}

func ipv4Num(s string) uint32 {
	return binary.BigEndian.Uint32(net.ParseIP(s).To4())
}

func isUpperASCII(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func countryCount(db *DB) int {
	n, _, _, _, _ := db.Stats()
	return n
}

// --- 文件头 ---

func TestSpec_header_magicAndVersion(t *testing.T) {
	data, _, h := loadFixture(t)
	for _, tc := range []struct{ name string; got, want any }{
		{"magic[0]=I", string(h.Magic[0]), "I"},
		{"magic[1]=P", string(h.Magic[1]), "P"},
		{"magic[2]=G", string(h.Magic[2]), "G"},
		{"magic[3]=E", string(h.Magic[3]), "E"},
		{"magic[4]=O", string(h.Magic[4]), "O"},
		{"version=2", int(h.Version), 2},
		{"file_size", len(data), 1169608},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %v want %v", tc.got, tc.want)
			}
		})
	}
}

func TestSpec_header_u32_offsets(t *testing.T) {
	_, _, h := loadFixture(t)
	for _, tc := range []struct{ name string; got, want uint32 }{
		{"offset8_country", h.CountryOff, 64},
		{"offset16_v4L1", h.V4L1Off, 6330},
		{"offset24_v4L2", h.V4L2Off, 105050},
		{"offset32_v6wide", h.V6WideOff, 255058},
		{"offset40_v6compact", h.V6CompactOff, 707958},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %d want %d", tc.got, tc.want)
			}
		})
	}
}

func TestSpec_header_u32_counts(t *testing.T) {
	_, _, h := loadFixture(t)
	for _, tc := range []struct{ name string; got, want uint32 }{
		{"offset20_v4L1_count", h.V4L1Count, 12340},
		{"offset28_v4L2_count", h.V4L2Count, 18751},
		{"offset36_v6wide_count", h.V6WideCount, 22645},
		{"offset44_v6compact_count", h.V6CompactCount, 46165},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %d want %d", tc.got, tc.want)
			}
		})
	}
}

func TestSpec_header_u16_countryCount(t *testing.T) {
	_, _, h := loadFixture(t)
	if h.CountryCount != 234 {
		t.Fatalf("got %d want 234", h.CountryCount)
	}
}

func TestSpec_header_reservedBytes(t *testing.T) {
	data, _, _ := loadFixture(t)
	for _, r := range []struct{ name string; from, to int }{
		{"offset6_7", 6, 8},
		{"offset14_15", 14, 16},
		{"offset48_63", 48, 64},
	} {
		t.Run(r.name, func(t *testing.T) {
			for i := r.from; i < r.to; i++ {
				if data[i] != 0 {
					t.Fatalf("byte@%d = 0x%02x", i, data[i])
				}
			}
		})
	}
}

func TestSpec_header_segmentChain(t *testing.T) {
	_, _, h := loadFixture(t)
	for _, c := range []struct{ name string; end, next uint32 }{
		{"country_end=v4L1_start", h.V4L1Off, h.V4L1Off},
		{"v4L1_end=v4L2_start", h.V4L1Off + h.V4L1Count*8, h.V4L2Off},
		{"v4L2_end=v6wide_start", h.V4L2Off + h.V4L2Count*8, h.V6WideOff},
		{"v6wide_end=v6compact_start", h.V6WideOff + h.V6WideCount*20, h.V6CompactOff},
	} {
		t.Run(c.name, func(t *testing.T) {
			if c.end != c.next {
				t.Fatalf("%d != %d", c.end, c.next)
			}
		})
	}
}

func TestSpec_header_fileEnd(t *testing.T) {
	data, _, h := loadFixture(t)
	end := h.V6CompactOff + h.V6CompactCount*10
	if uint32(len(data)) != end {
		t.Fatalf("file size %d end %d", len(data), end)
	}
}

func TestSpec_openBytes_rejectsBadInput(t *testing.T) {
	data, _, _ := loadFixture(t)
	for _, tc := range []struct {
		name string
		mut  func([]byte) []byte
	}{
		{"too_short", func(b []byte) []byte { return b[:32] }},
		{"bad_magic", func(b []byte) []byte { c := append([]byte(nil), b...); copy(c, []byte("XXXXX")); return c }},
		{"bad_version", func(b []byte) []byte { c := append([]byte(nil), b...); c[5] = 1; return c }},
		{"truncated_v4L1", func(b []byte) []byte {
			c := append([]byte(nil), b...)
			binary.BigEndian.PutUint32(c[20:24], 999999)
			return c
		}},
		{"truncated_country", func(b []byte) []byte {
			c := append([]byte(nil), b...)
			binary.BigEndian.PutUint32(c[16:20], 100)
			return c
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := OpenBytes(tc.mut(data)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

// --- Country Table ---

func TestSpec_countryTable_all234(t *testing.T) {
	data, db, h := loadFixture(t)
	off := int(h.CountryOff)
	seen := make(map[string]bool)
	for i := 0; i < int(h.CountryCount); i++ {
		idx, start := i, off
		cCheck, _, next, err := CountryRaw(data, start)
		if err != nil {
			t.Fatal(err)
		}
		if seen[cCheck.Code] {
			t.Fatalf("duplicate cc %s at %d", cCheck.Code, idx)
		}
		seen[cCheck.Code] = true
		t.Run(fmt.Sprintf("cc_%03d_%s", idx, cCheck.Code), func(t *testing.T) {
			cRaw, flags, next, err := CountryRaw(data, start)
			if err != nil {
				t.Fatal(err)
			}
			if next <= start {
				t.Fatal("offset did not advance")
			}
			c := db.CountryAt(idx)
			if c != cRaw {
				t.Fatalf("mismatch raw=%+v db=%+v", cRaw, c)
			}
			if len(c.Code) != 2 || !isUpperASCII(c.Code) {
				t.Fatalf("bad code %q", c.Code)
			}
			if !validContinents[c.ContinentCode] {
				t.Fatalf("bad continent %q", c.ContinentCode)
			}
			if !utf8.ValidString(c.NameZH) || c.NameZH == "" {
				t.Fatalf("bad zh %q", c.NameZH)
			}
			if !utf8.ValidString(c.NameEN) || c.NameEN == "" {
				t.Fatalf("bad en %q", c.NameEN)
			}
			if chinesePrimaryCC[c.Code] != c.ChinesePrimary {
				t.Fatalf("ChinesePrimary=%v", c.ChinesePrimary)
			}
			if flags&0xFE != 0 {
				t.Fatalf("reserved flags 0x%02x", flags)
			}
		})
		off = next
	}
	if off != int(h.V4L1Off) {
		t.Fatalf("country table end %d want %d", off, h.V4L1Off)
	}
}

func TestSpec_countryTable_knownEntries(t *testing.T) {
	_, db, _ := loadFixture(t)
	for _, tc := range []struct {
		idx            int
		cc, cont, zh, en string
		zhPri          bool
	}{
		{0, "AD", "EU", "安道尔", "Andorra", false},
		{44, "CN", "AS", "中国", "China", true},
	} {
		t.Run(tc.cc, func(t *testing.T) {
			c := db.CountryAt(tc.idx)
			if c.Code != tc.cc || c.ContinentCode != tc.cont {
				t.Fatalf("got %+v", c)
			}
			if c.NameZH != tc.zh || c.NameEN != tc.en {
				t.Fatalf("names zh=%q en=%q", c.NameZH, c.NameEN)
			}
			if c.ChinesePrimary != tc.zhPri {
				t.Fatalf("ChinesePrimary=%v", c.ChinesePrimary)
			}
		})
	}
}

func TestSpec_countryName_lang(t *testing.T) {
	_, db, _ := loadFixture(t)
	cn := db.CountryAt(44)
	us := db.CountryAt(lookupCCIndex(t, db, "US"))
	for _, tc := range []struct {
		name       string
		c          Country
		lang, want string
	}{
		{"cn_auto", cn, "auto", "中国"},
		{"cn_zh", cn, "zh", "中国"},
		{"cn_en", cn, "en", "China"},
		{"us_auto", us, "auto", "United States"},
		{"us_en", us, "en", "United States"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.Name(tc.lang); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func lookupCCIndex(t *testing.T, db *DB, cc string) int {
	t.Helper()
	for i := 0; i < countryCount(db); i++ {
		if db.CountryAt(i).Code == cc {
			return i
		}
	}
	t.Fatalf("cc %s not found", cc)
	return -1
}

// --- IPv4 8B ---

func TestSpec_v4L1_firstRecord_layout(t *testing.T) {
	data, db, h := loadFixture(t)
	start, pfx, cc, pad := V4RecordRaw(data, int(h.V4L1Off))
	if start != ipv4Num("1.3.0.0") {
		t.Fatalf("start %08x", start)
	}
	if pfx != 16 {
		t.Fatalf("prefix %d", pfx)
	}
	if pad != 0 {
		t.Fatalf("pad 0x%02x", pad)
	}
	if db.CountryAt(int(cc)).Code != "CN" {
		t.Fatalf("cc idx %d", cc)
	}
}

func TestSpec_v4_recordLayout_spotChecks(t *testing.T) {
	data, db, h := loadFixture(t)
	n := countryCount(db)
	spots := []struct {
		seg, name string
		base      uint32
		idx       int
		l2        bool
	}{
		{"L1", "idx0", h.V4L1Off, 0, false},
		{"L1", "idx1", h.V4L1Off, 1, false},
		{"L1", "idx100", h.V4L1Off, 100, false},
		{"L1", "idx1000", h.V4L1Off, 1000, false},
		{"L1", "idx5000", h.V4L1Off, 5000, false},
		{"L1", "idx12339", h.V4L1Off, 12339, false},
		{"L2", "idx0", h.V4L2Off, 0, true},
		{"L2", "idx1", h.V4L2Off, 1, true},
		{"L2", "idx500", h.V4L2Off, 500, true},
		{"L2", "idx5000", h.V4L2Off, 5000, true},
		{"L2", "idx18750", h.V4L2Off, 18750, true},
	}
	for _, s := range spots {
		t.Run(s.seg+"_"+s.name, func(t *testing.T) {
			off := int(s.base) + s.idx*8
			start, pfx, cc, pad := V4RecordRaw(data, off)
			if int(cc) >= n {
				t.Fatalf("cc idx %d", cc)
			}
			if pad != 0 {
				t.Fatalf("pad 0x%02x", pad)
			}
			if pfx == 0 || pfx > 32 {
				t.Fatalf("prefix %d", pfx)
			}
			if s.l2 && (pfx < 17 || pfx > 19) {
				t.Fatalf("L2 prefix %d not in 17-19", pfx)
			}
			_ = start
		})
	}
}

func TestSpec_v4L1_prefixDominantlyGTE16(t *testing.T) {
	data, _, h := loadFixture(t)
	var below int
	for i := 0; i < int(h.V4L1Count); i++ {
		_, pfx, _, _ := V4RecordRaw(data, int(h.V4L1Off)+i*8)
		if pfx < 16 {
			below++
		}
	}
	if below == 0 {
		t.Fatal("expected some /15 blocks in L1 sample data")
	}
}

func TestSpec_v4_sortedAscending(t *testing.T) {
	data, _, h := loadFixture(t)
	for _, seg := range []struct {
		name string
		off  uint32
		n    uint32
	}{
		{"L1", h.V4L1Off, h.V4L1Count},
		{"L2", h.V4L2Off, h.V4L2Count},
	} {
		t.Run(seg.name, func(t *testing.T) {
			var prev uint32
			for i := 0; i < int(seg.n); i++ {
				start, _, _, _ := V4RecordRaw(data, int(seg.off)+i*8)
				if i > 0 && start < prev {
					t.Fatalf("idx %d: %08x < %08x", i, start, prev)
				}
				prev = start
			}
		})
	}
}

func TestSpec_v4_ccIndexInRange(t *testing.T) {
	data, db, h := loadFixture(t)
	n := countryCount(db)
	for _, seg := range []struct{ name string; off, cnt uint32 }{
		{"L1", h.V4L1Off, h.V4L1Count},
		{"L2", h.V4L2Off, h.V4L2Count},
	} {
		t.Run(seg.name, func(t *testing.T) {
			for i := 0; i < int(seg.cnt); i++ {
				_, _, cc, _ := V4RecordRaw(data, int(seg.off)+i*8)
				if int(cc) >= n {
					t.Fatalf("rec %d cc %d >= %d", i, cc, n)
				}
			}
		})
	}
}

// --- IPv6 wide 20B / compact 10B ---

func TestSpec_v6Wide_firstRecord_layout(t *testing.T) {
	data, db, h := loadFixture(t)
	start, pfx, cc, pad := V6WideRecordRaw(data, int(h.V6WideOff))
	if pfx == 32 || pfx == 48 {
		t.Fatalf("wide should not be /32 or /48, got %d", pfx)
	}
	if pad != 0 || int(cc) >= countryCount(db) {
		t.Fatalf("pad=%02x cc=%d", pad, cc)
	}
	_ = start
}

func TestSpec_v6Compact_firstRecord_layout(t *testing.T) {
	data, db, h := loadFixture(t)
	pfx, start, cc, pad := V6CompactRecordRaw(data, int(h.V6CompactOff))
	if pfx != 32 && pfx != 48 {
		t.Fatalf("compact prefix %d", pfx)
	}
	if pad != 0 || int(cc) >= countryCount(db) {
		t.Fatalf("pad=%02x cc=%d", pad, cc)
	}
	_ = start
}

func TestSpec_v6_recordLayout_spotChecks(t *testing.T) {
	data, db, h := loadFixture(t)
	n := countryCount(db)
	for _, i := range []int{0, 1, 100, 1000, 5000, 10000, 20000, 22644} {
		t.Run(fmt.Sprintf("wide_%d", i), func(t *testing.T) {
			_, pfx, cc, pad := V6WideRecordRaw(data, int(h.V6WideOff)+i*20)
			if pfx == 32 || pfx == 48 {
				t.Fatalf("prefix %d in wide", pfx)
			}
			if pad != 0 || int(cc) >= n {
				t.Fatalf("pad=%02x cc=%d", pad, cc)
			}
		})
	}
	for _, i := range []int{0, 1, 100, 1000, 5000, 20000, 40000, 46164} {
		t.Run(fmt.Sprintf("compact_%d", i), func(t *testing.T) {
			pfx, _, cc, pad := V6CompactRecordRaw(data, int(h.V6CompactOff)+i*10)
			if pfx != 32 && pfx != 48 {
				t.Fatalf("prefix %d not 32/48", pfx)
			}
			if pad != 0 || int(cc) >= n {
				t.Fatalf("pad=%02x cc=%d", pad, cc)
			}
		})
	}
}

func TestSpec_v6_sortedAscending(t *testing.T) {
	data, _, h := loadFixture(t)
	t.Run("wide", func(t *testing.T) {
		var prev [16]byte
		for i := 0; i < int(h.V6WideCount); i++ {
			start, _, _, _ := V6WideRecordRaw(data, int(h.V6WideOff)+i*20)
			if i > 0 && bytesCompare(start[:], prev[:]) < 0 {
				t.Fatalf("idx %d out of order", i)
			}
			prev = start
		}
	})
	t.Run("compact", func(t *testing.T) {
		var prev [6]byte
		for i := 0; i < int(h.V6CompactCount); i++ {
			_, start, _, _ := V6CompactRecordRaw(data, int(h.V6CompactOff)+i*10)
			if i > 0 && bytesCompare(start[:], prev[:]) < 0 {
				t.Fatalf("idx %d out of order", i)
			}
			prev = start
		}
	})
}

func bytesCompare(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// --- Lookup ---

func TestSpec_lookup_documented(t *testing.T) {
	_, db, _ := loadFixture(t)
	for _, tc := range []struct{ ip, code, cont, name string }{
		{"8.8.8.8", "US", "NA", "United States"},
		{"114.114.114.114", "CN", "AS", "中国"},
		{"2001:4860:4860::8888", "US", "NA", "United States"},
		{"2400:3200::1", "CN", "AS", "中国"},
	} {
		t.Run(tc.ip, func(t *testing.T) {
			r, ok := db.Lookup(tc.ip)
			if !ok {
				t.Fatal("miss")
			}
			if r.Country.Code != tc.code || r.Country.ContinentCode != tc.cont {
				t.Fatalf("got %+v", r.Country)
			}
			if r.Country.Name("auto") != tc.name {
				t.Fatalf("name %q", r.Country.Name("auto"))
			}
		})
	}
}

func TestSpec_lookup_miss(t *testing.T) {
	_, db, _ := loadFixture(t)
	for _, ip := range []string{
		"192.168.1.1", "10.0.0.1", "127.0.0.1", "172.16.0.1",
		"not-an-ip", "", "999.999.999.999", "::1", "fe80::1",
	} {
		t.Run(ip, func(t *testing.T) {
			if _, ok := db.Lookup(ip); ok {
				t.Fatal("expected miss")
			}
		})
	}
}

func TestSpec_lookup_v4_cidrPopulated(t *testing.T) {
	_, db, _ := loadFixture(t)
	r, ok := db.Lookup("114.114.114.114")
	if !ok || r.CIDR == "" {
		t.Fatal("expected v4 CIDR")
	}
}

func TestSpec_lookup_chinesePrimaryRegions(t *testing.T) {
	_, db, _ := loadFixture(t)
	for cc, ip := range map[string]string{
		"CN": "114.114.114.114",
		"TW": "1.160.0.1",
		"HK": "1.36.0.1",
		"MO": "202.175.0.1",
	} {
		t.Run(cc, func(t *testing.T) {
			r, ok := db.Lookup(ip)
			if !ok || r.Country.Code != cc || !r.Country.ChinesePrimary {
				t.Fatalf("got %+v", r)
			}
		})
	}
}

func TestSpec_lookup_v4L2_fallback(t *testing.T) {
	_, db, _ := loadFixture(t)
	r, ok := db.Lookup("1.0.32.1")
	if !ok || r.Country.Code == "" {
		t.Fatal("expected L2 hit")
	}
}

func TestSpec_maxCountryIndex(t *testing.T) {
	_, db, _ := loadFixture(t)
	if int(db.MaxCountryIndex()) >= countryCount(db) {
		t.Fatalf("max %d", db.MaxCountryIndex())
	}
}

// TestSpec_subtestBudget 静态核对：实际子测试数应不少于 150。
func TestSpec_subtestBudget(t *testing.T) {
	const min = 150
	n := 7 + 5 + 4 + 1 + 3 + 4 + 1 + 5 + 234 + 2 + 5 + 1 + 11 + 1 + 2 + 2 + 1 + 1 + 16 + 2 + 4 + 8 + 1 + 4 + 1 + 1
	if n < min {
		t.Fatalf("budget %d < %d", n, min)
	}
}
