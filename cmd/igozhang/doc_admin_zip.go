package main

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/igokorea/igozhang/internal/docindex"
	"github.com/igokorea/igozhang/internal/markdown"
)

const (
	zipImportArchiveMax = 128 << 20
	zipImportMDMax      = 8 << 20
	zipImportPicMax     = 32 << 20
	zipImportMetaMax    = 1 << 20
	zipImportSSHPassMax = 32 << 10
	zipImportIgoConfMax = 512 << 10
	zipImportEngMax       = 4 << 20
	zipImportBlacklistMax = 64 << 10
	zipImportWhitelistMax = 64 << 10
)

// zipNormalizeEntry rejects absolute paths and zip-slip (../).
func zipNormalizeEntry(name string) (string, bool) {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, seg := range strings.Split(name, "/") {
		if seg == ".." {
			return "", false
		}
	}
	name = path.Clean(name)
	if name == "." || name == ".." {
		return "", false
	}
	if strings.HasPrefix(name, "/") {
		return "", false
	}
	return name, true
}

func zipLegacyEntry(entry string) bool {
	return entry == "doc_info.md" || entry == "doc_views.md"
}

func zipAssetsDest(assetsRoot, entry string) (string, bool) {
	switch {
	case strings.HasPrefix(entry, "md/"):
		return safeJoin(assetsRoot, entry)
	case strings.HasPrefix(entry, "doc-pic/"):
		return safeJoin(assetsRoot, entry)
	case strings.HasPrefix(entry, "eng/"):
		return safeJoin(assetsRoot, entry)
	default:
		return "", false
	}
}

func zipReadLimit(zf *zip.File, max int64) ([]byte, error) {
	if zf.UncompressedSize64 > uint64(max) {
		return nil, fmt.Errorf("entry exceeds %d bytes", max)
	}
	rc, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	raw, err := io.ReadAll(io.LimitReader(rc, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > max {
		return nil, fmt.Errorf("entry exceeds %d bytes", max)
	}
	return raw, nil
}

func zipExtractFile(zf *zip.File, dest string, max int64) error {
	raw, err := zipReadLimit(zf, max)
	if err != nil {
		return err
	}
	return zipWriteFile(dest, raw, 0o644)
}

func zipWriteFile(dest string, raw []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, raw, perm)
}

type secret07ZipImportSpec struct {
	dest string
	max  int64
	perm os.FileMode
	kind string // banner | favicon | texas | link.conf | ssh.pass | igo.conf | common-commands | ip-blacklist | ip-whitelist | user-access-bundle
}

// zipSecret07BackupEntries 与 appendSecret07ZipOptional 导出项一致（导入须全部可恢复）。
// common-commands.md：与 link.conf 相同，secret07ReadFileOrMissing 判定有定制才导出；导入写盘后 /ops 即生效。

func secret07ZipImportTarget(dataRoot, entry string) (secret07ZipImportSpec, bool) {
	if !strings.HasPrefix(entry, "secret07/") {
		return secret07ZipImportSpec{}, false
	}
	base := path.Base(entry)
	switch entry {
	case "secret07/link.conf":
		return secret07ZipImportSpec{
			dest: linkConfPath(dataRoot),
			max:  zipImportMetaMax,
			perm: 0o644,
			kind: "link.conf",
		}, true
	case "secret07/ssh.pass":
		return secret07ZipImportSpec{
			dest: sshPassPath(dataRoot),
			max:  zipImportSSHPassMax,
			perm: 0o600,
			kind: "ssh.pass",
		}, true
	case "secret07/igo.conf":
		return secret07ZipImportSpec{
			dest: opsConfigPath(dataRoot),
			max:  zipImportIgoConfMax,
			perm: 0o644,
			kind: "igo.conf",
		}, true
	case "secret07/common-commands.md":
		return secret07ZipImportSpec{
			dest: commonCommandsPath(dataRoot),
			max:  zipImportMetaMax,
			perm: 0o644,
			kind: "common-commands",
		}, true
	case "secret07/ip-blacklist.conf":
		return secret07ZipImportSpec{
			dest: ipBlacklistPath(dataRoot),
			max:  zipImportBlacklistMax,
			perm: 0o600,
			kind: "ip-blacklist",
		}, true
	case "secret07/ip-whitelist.conf":
		return secret07ZipImportSpec{
			dest: ipWhitelistPath(dataRoot),
			max:  zipImportWhitelistMax,
			perm: 0o600,
			kind: "ip-whitelist",
		}, true
	case "secret07/user-access-bundle.json":
		return secret07ZipImportSpec{
			dest: filepath.Join(secret07Dir(dataRoot), "user-access-bundle.json"),
			max:  zipImportAccessBundleMax,
			perm: 0o644,
			kind: "user-access-bundle",
		}, true
	}
	if _, ok := splitBannerName(base); ok {
		return secret07ZipImportSpec{
			dest: filepath.Join(secret07Dir(dataRoot), base),
			max:  siteBannerMaxBytes,
			perm: 0o644,
			kind: "banner",
		}, true
	}
	if stem, _, ok := splitFaviconName(base); ok && (stem == "favicon" || stem == "fav") {
		return secret07ZipImportSpec{
			dest: filepath.Join(secret07Dir(dataRoot), base),
			max:  siteFaviconMaxBytes,
			perm: 0o644,
			kind: "favicon",
		}, true
	}
	if _, _, ok := splitTexasCardBackName(base); ok {
		return secret07ZipImportSpec{
			dest: filepath.Join(secret07Dir(dataRoot), base),
			max:  texasCardBackMaxBytes,
			perm: 0o644,
			kind: "texas",
		}, true
	}
	return secret07ZipImportSpec{}, false
}

func clearSecret07BannerFiles(dataRoot string) error {
	return clearSecret07Matching(dataRoot, func(name string) bool {
		_, ok := splitBannerName(name)
		return ok
	})
}

func clearSecret07FaviconFiles(dataRoot string) error {
	return clearSecret07Matching(dataRoot, func(name string) bool {
		_, _, ok := splitFaviconName(name)
		return ok
	})
}

func clearSecret07TexasCardBackFiles(dataRoot string) error {
	return clearSecret07Matching(dataRoot, func(name string) bool {
		_, _, ok := splitTexasCardBackName(name)
		return ok
	})
}

func clearSecret07Matching(dataRoot string, match func(name string) bool) error {
	dir := secret07Dir(dataRoot)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !match(e.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func importSecret07FromZip(dataRoot string, items []struct {
	entry string
	zf    *zip.File
	spec  secret07ZipImportSpec
}, importErrs, importWarnings *[]string) map[string]bool {
	imported := make(map[string]bool)
	if dataRoot == "" || len(items) == 0 {
		return imported
	}
	if err := os.MkdirAll(secret07Dir(dataRoot), 0o755); err != nil {
		*importErrs = append(*importErrs, "secret07/: "+err.Error())
		return imported
	}
	var clearBanner, clearFavicon, clearTexas bool
	for _, item := range items {
		switch item.spec.kind {
		case "banner":
			clearBanner = true
		case "favicon":
			clearFavicon = true
		case "texas":
			clearTexas = true
		}
	}
	if clearBanner {
		if err := clearSecret07BannerFiles(dataRoot); err != nil {
			*importErrs = append(*importErrs, "secret07/banner.*: clear "+err.Error())
		}
	}
	if clearFavicon {
		if err := clearSecret07FaviconFiles(dataRoot); err != nil {
			*importErrs = append(*importErrs, "secret07/favicon.*: clear "+err.Error())
		}
	}
	if clearTexas {
		if err := clearSecret07TexasCardBackFiles(dataRoot); err != nil {
			*importErrs = append(*importErrs, "secret07/texas.*: clear "+err.Error())
		}
	}
	for _, item := range items {
		raw, err := zipReadLimit(item.zf, item.spec.max)
		if err != nil {
			*importErrs = append(*importErrs, item.entry+": "+err.Error())
			continue
		}
		if len(raw) == 0 {
			*importErrs = append(*importErrs, item.entry+": empty")
			continue
		}
		if item.spec.kind == "igo.conf" {
			raw, err = normalizeIgoConfImportBytes(raw)
			if err != nil {
				*importErrs = append(*importErrs, item.entry+": "+err.Error())
				continue
			}
		}
		if item.spec.kind == "ip-blacklist" {
			raw, err = normalizeIPBlacklistImportBytes(raw)
			if err != nil {
				*importErrs = append(*importErrs, item.entry+": "+err.Error())
				continue
			}
		}
		if item.spec.kind == "ip-whitelist" {
			raw, err = normalizeIPWhitelistImportBytes(raw)
			if err != nil {
				*importErrs = append(*importErrs, item.entry+": "+err.Error())
				continue
			}
		}
		if item.spec.kind == "user-access-bundle" {
			warn, err := importUserAccessBundle(dataRoot, raw)
			if err != nil {
				*importErrs = append(*importErrs, item.entry+": "+err.Error())
				continue
			}
			for _, w := range warn {
				*importWarnings = append(*importWarnings, item.entry+": "+w)
			}
			imported[item.spec.kind] = true
			continue
		}
		if err := zipWriteFile(item.spec.dest, raw, item.spec.perm); err != nil {
			*importErrs = append(*importErrs, item.entry+": "+err.Error())
			continue
		}
		imported[item.spec.kind] = true
	}
	return imported
}

func postSecret07ZipImport(dataRoot string, imported map[string]bool) {
	if dataRoot == "" || len(imported) == 0 {
		return
	}
	if imported["igo.conf"] {
		if err := loadOpsConfig(dataRoot); err != nil {
			log.Printf("ops-config: reload after zip import: %v", err)
		}
	}
	if imported["ip-blacklist"] {
		reloadIPBlacklist(dataRoot)
	}
	if imported["ip-whitelist"] {
		reloadIPWhitelist(dataRoot)
	}
	// user-access-bundle：importUserAccessBundle 内清理窗口、写盘、同步 last-day、恢复 insights 并清空内存看板
}

func zipAddBytes(zw *zip.Writer, name string, data []byte) error {
	name = strings.ReplaceAll(name, "\\", "/")
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func zipAddTree(zw *zip.Writer, srcDir, prefix string) error {
	st, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", srcDir)
	}
	return filepath.Walk(srcDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, p)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return zipAddBytes(zw, path.Join(prefix, filepath.ToSlash(rel)), raw)
	})
}

func zipAddSecret07Raster(zw *zip.Writer, srcPath, fileName string) error {
	raw, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return zipAddBytes(zw, "secret07/"+fileName, raw)
}

func appendSecret07ZipOptional(dataRoot string, zw *zip.Writer, zipErrs *[]string, accessBundle []byte) {
	if raw := readLinkConfExportBytes(dataRoot); len(raw) > 0 {
		if err := zipAddBytes(zw, "secret07/link.conf", raw); err != nil {
			*zipErrs = append(*zipErrs, "secret07/link.conf: "+err.Error())
		}
	}
	if raw := readIgoConfExportBytes(dataRoot); len(raw) > 0 {
		if err := zipAddBytes(zw, "secret07/igo.conf", raw); err != nil {
			*zipErrs = append(*zipErrs, "secret07/igo.conf: "+err.Error())
		}
	}
	if raw := readSSHPassExportBytes(dataRoot); len(raw) > 0 {
		if err := zipAddBytes(zw, "secret07/ssh.pass", raw); err != nil {
			*zipErrs = append(*zipErrs, "secret07/ssh.pass: "+err.Error())
		}
	}
	if raw := readCommonCommandsExportBytes(dataRoot); len(raw) > 0 {
		if err := zipAddBytes(zw, "secret07/common-commands.md", raw); err != nil {
			*zipErrs = append(*zipErrs, "secret07/common-commands.md: "+err.Error())
		}
	}
	if c := resolveSecret07Banner(dataRoot); c != nil {
		if err := zipAddSecret07Raster(zw, c.Path, c.FileName); err != nil {
			*zipErrs = append(*zipErrs, "secret07/"+c.FileName+": "+err.Error())
		}
	}
	if c := resolveSecret07Favicon(dataRoot); c != nil {
		if err := zipAddSecret07Raster(zw, c.Path, c.FileName); err != nil {
			*zipErrs = append(*zipErrs, "secret07/"+c.FileName+": "+err.Error())
		}
	}
	if c := resolveSecret07TexasCardBack(dataRoot); c != nil {
		if err := zipAddSecret07Raster(zw, c.Path, c.FileName); err != nil {
			*zipErrs = append(*zipErrs, "secret07/"+c.FileName+": "+err.Error())
		}
	}
	if raw := readIPBlacklistExportBytes(dataRoot); len(raw) > 0 {
		if err := zipAddBytes(zw, "secret07/ip-blacklist.conf", raw); err != nil {
			*zipErrs = append(*zipErrs, "secret07/ip-blacklist.conf: "+err.Error())
		}
	}
	if raw := readIPWhitelistExportBytes(dataRoot); len(raw) > 0 {
		if err := zipAddBytes(zw, "secret07/ip-whitelist.conf", raw); err != nil {
			*zipErrs = append(*zipErrs, "secret07/ip-whitelist.conf: "+err.Error())
		}
	}
	if len(accessBundle) > 0 {
		if err := zipAddBytes(zw, userAccessBundleZipEntry, accessBundle); err != nil {
			*zipErrs = append(*zipErrs, userAccessBundleZipEntry+": "+err.Error())
		}
	}
}

func buildDocExportZip(dataRoot, assets string, st *docindex.Store) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeDocExportZip(&buf, dataRoot, assets, st, false); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeDocExportZip(w io.Writer, dataRoot, assets string, st *docindex.Store, prepareAccessLogs bool) error {
	all, err := st.ListAll()
	if err != nil {
		return fmt.Errorf("list articles: %w", err)
	}
	mdRoot := filepath.Join(assets, "md")
	missingMd := make(map[string]struct{})
	for _, m := range all {
		if _, err := os.Stat(filepath.Join(mdRoot, m.Slug+".md")); err != nil {
			if os.IsNotExist(err) {
				missingMd[m.Slug] = struct{}{}
				continue
			}
			return err
		}
	}

	zw := zip.NewWriter(w)
	var zipErrs []string

	exportedMeta := len(all) - len(missingMd)
	accessBundle, accessBundleErr := buildUserAccessBundleBytesOpts(dataRoot, prepareAccessLogs)
	if accessBundleErr != nil {
		_ = zw.Close()
		return fmt.Errorf("access bundle: %w", accessBundleErr)
	}
	manifest := map[string]any{
		"version":    2,
		"exportedAt": time.Now().Format(time.RFC3339),
		"count":      exportedMeta,
	}
	if len(accessBundle) > 0 {
		manifest["accessBundle"] = userAccessBundleZipEntry
		if sm := peekBundleStatsManifest(accessBundle); sm != nil {
			manifest["accessBundleStats"] = sm
		}
	}
	if len(missingMd) > 0 {
		slugs := make([]string, 0, len(missingMd))
		for slug := range missingMd {
			slugs = append(slugs, slug)
		}
		manifest["missingMd"] = slugs
	}
	mb, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		_ = zw.Close()
		return err
	}
	if err := zipAddBytes(zw, "manifest.json", mb); err != nil {
		zipErrs = append(zipErrs, "manifest.json: "+err.Error())
	}
	for _, m := range all {
		if _, skip := missingMd[m.Slug]; skip {
			continue
		}
		mb, err := json.Marshal(m)
		if err != nil {
			zipErrs = append(zipErrs, "meta/"+m.Slug+".json: marshal")
			continue
		}
		if err := zipAddBytes(zw, "meta/"+m.Slug+".json", mb); err != nil {
			zipErrs = append(zipErrs, "meta/"+m.Slug+".json: "+err.Error())
		}
	}
	if err := zipAddTree(zw, mdRoot, "md"); err != nil {
		zipErrs = append(zipErrs, "md/: "+err.Error())
	}
	if err := zipAddTree(zw, filepath.Join(assets, "doc-pic"), "doc-pic"); err != nil {
		zipErrs = append(zipErrs, "doc-pic/: "+err.Error())
	}
	if err := zipAddTree(zw, filepath.Join(assets, assetsEngRelDir), assetsEngRelDir); err != nil {
		zipErrs = append(zipErrs, assetsEngRelDir+"/: "+err.Error())
	}
	appendSecret07ZipOptional(dataRoot, zw, &zipErrs, accessBundle)
	if err := zw.Close(); err != nil {
		return err
	}
	if len(zipErrs) > 0 {
		return fmt.Errorf("export incomplete: %s", strings.Join(zipErrs, "; "))
	}
	return nil
}

func clearAssetFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func importDocZip(dataRoot, assets string, st *docindex.Store, zr *zip.Reader) (int, []string, error) {
	type extractItem struct {
		entry string
		zf    *zip.File
	}

	var (
		importErrs      []string
		importWarnings  []string
		metaPending   []docindex.Meta
		metaInZip     = make(map[string]struct{})
		toExtract     []extractItem
		secret07Items []struct {
			entry string
			zf    *zip.File
			spec  secret07ZipImportSpec
		}
		slugs = make(map[string]struct{})
	)

	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		entry, ok := zipNormalizeEntry(zf.Name)
		if !ok {
			importErrs = append(importErrs, zf.Name+": invalid path")
			continue
		}
		switch {
		case strings.HasPrefix(entry, "meta/") && strings.HasSuffix(entry, ".json"):
			raw, err := zipReadLimit(zf, zipImportMetaMax)
			if err != nil {
				importErrs = append(importErrs, entry+": "+err.Error())
				continue
			}
			var m docindex.Meta
			if err := json.Unmarshal(raw, &m); err != nil {
				importErrs = append(importErrs, entry+": bad json")
				continue
			}
			slug := strings.TrimSuffix(path.Base(entry), ".json")
			if m.Slug == "" {
				m.Slug = slug
			}
			if m.Slug != slug {
				importErrs = append(importErrs, entry+": slug mismatch")
				continue
			}
			if !docindex.ValidSlug(m.Slug) {
				importErrs = append(importErrs, entry+": invalid slug")
				continue
			}
			metaPending = append(metaPending, m)
			metaInZip[m.Slug] = struct{}{}
			slugs[m.Slug] = struct{}{}
		case strings.HasPrefix(entry, "md/") && strings.HasSuffix(entry, ".md"):
			slug := strings.TrimSuffix(path.Base(entry), ".md")
			if docindex.ValidSlug(slug) {
				slugs[slug] = struct{}{}
			}
			toExtract = append(toExtract, extractItem{entry: entry, zf: zf})
		case strings.HasPrefix(entry, "doc-pic/"):
			toExtract = append(toExtract, extractItem{entry: entry, zf: zf})
		case strings.HasPrefix(entry, "eng/"):
			toExtract = append(toExtract, extractItem{entry: entry, zf: zf})
		default:
			if spec, ok := secret07ZipImportTarget(dataRoot, entry); ok {
				secret07Items = append(secret07Items, struct {
					entry string
					zf    *zip.File
					spec  secret07ZipImportSpec
				}{entry: entry, zf: zf, spec: spec})
				continue
			}
			if entry == "manifest.json" || zipLegacyEntry(entry) {
				continue
			}
			continue
		}
	}
	if len(slugs) == 0 {
		return 0, nil, fmt.Errorf("empty import: no articles in zip")
	}

	importedAt := time.Now()
	mdSkipped := make(map[string]bool)
	importedEng := false

	for _, item := range toExtract {
		dest, ok := zipAssetsDest(assets, item.entry)
		if !ok {
			continue
		}
		max := zipImportMDMax
		switch {
		case strings.HasPrefix(item.entry, "doc-pic/"):
			max = zipImportPicMax
		case strings.HasPrefix(item.entry, "eng/"):
			max = zipImportEngMax
		}
		if strings.HasPrefix(item.entry, "md/") && strings.HasSuffix(item.entry, ".md") {
			zipRaw, err := zipReadLimit(item.zf, int64(max))
			if err != nil {
				importErrs = append(importErrs, item.entry+": "+err.Error())
				continue
			}
			localRaw, readErr := os.ReadFile(dest)
			if readErr == nil && bytes.Equal(localRaw, zipRaw) {
				slug := strings.TrimSuffix(path.Base(item.entry), ".md")
				if docindex.ValidSlug(slug) {
					mdSkipped[slug] = true
				}
				continue
			}
			if err := zipWriteFile(dest, zipRaw, 0o644); err != nil {
				importErrs = append(importErrs, item.entry+": "+err.Error())
			}
			continue
		}
		if err := zipExtractFile(item.zf, dest, int64(max)); err != nil {
			importErrs = append(importErrs, item.entry+": "+err.Error())
			continue
		}
		if strings.HasPrefix(item.entry, "eng/") {
			importedEng = true
		}
	}

	for _, m := range metaPending {
		body, err := st.ReadBody(m.Slug)
		if err != nil {
			// 无正文无法恢复索引；跳过（旧版 ZIP 可能含 orphan meta）。
			continue
		}
		if m.Title == "" {
			m.Title = markdown.ResolveArticleTitle("", m.Slug, body)
		}
		if m.Title == "" {
			m.Title = m.Slug
		}
		if strings.TrimSpace(m.Author) == "" {
			m.Author = docindex.DefaultAuthor
		}
		var impErr error
		if mdSkipped[m.Slug] {
			impErr = st.ImportArticleMeta(m)
		} else {
			impErr = st.ImportArticle(m, body)
		}
		if impErr != nil {
			importErrs = append(importErrs, "meta/"+m.Slug+".json: "+impErr.Error())
			continue
		}
	}

	if err := importIndexZipMdWithoutMeta(assets, st, slugs, metaInZip, importedAt, &importErrs); err != nil {
		return 0, nil, err
	}
	if err := importIndexLocalMdNotInZip(assets, st, slugs, importedAt, &importErrs); err != nil {
		return 0, nil, err
	}

	if err := st.Reconcile(); err != nil {
		return 0, nil, fmt.Errorf("reconcile: %w", err)
	}

	importSecret07Imported := importSecret07FromZip(dataRoot, secret07Items, &importErrs, &importWarnings)
	postSecret07ZipImport(dataRoot, importSecret07Imported)
	if importedEng {
		postAssetsEngZipImport(dataRoot)
	}

	all, err := st.ListAll()
	if err != nil {
		importErrs = append(importErrs, "list: "+err.Error())
	}
	count := len(all)
	if len(importErrs) > 0 {
		return count, importWarnings, fmt.Errorf("%s", strings.Join(importErrs, "; "))
	}
	return count, importWarnings, nil
}

func importRandomViews1to5() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(5))
	if err != nil {
		return 3
	}
	return n.Int64() + 1
}

// importIndexZipMdWithoutMeta 为 ZIP 含 md/ 但无 meta/ 的有效 slug 补建索引（导出时磁盘 orphan md 等场景）。
func importIndexZipMdWithoutMeta(assets string, st *docindex.Store, zipSlugs, metaInZip map[string]struct{}, importedAt time.Time, importErrs *[]string) error {
	for slug := range zipSlugs {
		if !docindex.ValidSlug(slug) {
			continue
		}
		if _, hasMeta := metaInZip[slug]; hasMeta {
			continue
		}
		ok, err := st.Exists(slug)
		if err != nil {
			*importErrs = append(*importErrs, "zip md "+slug+": "+err.Error())
			continue
		}
		if ok {
			continue
		}
		body, err := st.ReadBody(slug)
		if err != nil {
			*importErrs = append(*importErrs, "zip md "+slug+": "+err.Error())
			continue
		}
		m := docindex.Meta{
			Slug:        slug,
			Title:       markdown.ResolveArticleTitle("", slug, body),
			Author:      docindex.DefaultAuthor,
			PublishedAt: importedAt,
			Views:       importRandomViews1to5(),
		}
		if m.Title == "" {
			m.Title = slug
		}
		if err := st.ImportArticleMeta(m); err != nil {
			*importErrs = append(*importErrs, "zip md "+slug+": "+err.Error())
		}
	}
	return nil
}

// importIndexLocalMdNotInZip 为 ZIP 未包含的本地 md 保留文件并补建索引（无 DB 行时按导入时刻登记）。
func importIndexLocalMdNotInZip(assets string, st *docindex.Store, zipSlugs map[string]struct{}, importedAt time.Time, importErrs *[]string) error {
	mdDir := filepath.Join(assets, "md")
	entries, err := os.ReadDir(mdDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		if !docindex.ValidSlug(slug) {
			continue
		}
		if _, inZip := zipSlugs[slug]; inZip {
			continue
		}
		ok, err := st.Exists(slug)
		if err != nil {
			*importErrs = append(*importErrs, "local "+slug+": "+err.Error())
			continue
		}
		if ok {
			continue
		}
		body, err := st.ReadBody(slug)
		if err != nil {
			*importErrs = append(*importErrs, "local "+slug+": "+err.Error())
			continue
		}
		m := docindex.Meta{
			Slug:        slug,
			Title:       markdown.ResolveArticleTitle("", slug, body),
			Author:      docindex.DefaultAuthor,
			PublishedAt: importedAt,
			Views:       importRandomViews1to5(),
		}
		if m.Title == "" {
			m.Title = slug
		}
		if err := st.ImportArticleMeta(m); err != nil {
			*importErrs = append(*importErrs, "local "+slug+": "+err.Error())
		}
	}
	return nil
}
