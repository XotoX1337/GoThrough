// Package configstore manages the catalog of community walkthrough configs.
//
// Only the catalog (index.json) is embedded in the binary; the walkthrough
// YAMLs themselves are downloaded on demand from raw.githubusercontent.com (a
// CDN, NOT the rate-limited api.github.com) and cached on disk next to
// progress.json/settings.json. At startup the index is refreshed once from the
// CDN (falling back to the embedded copy when offline); picking a game downloads
// all of its chapters into the cache so chapter loads — and `next:` hand-offs —
// work instantly and offline. A per-chapter sha256 hash in the index lets the
// app detect and auto-update changed chapters without a version bump.
package configstore

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
)

//go:embed configs/index.json
var embeddedIndex []byte

// rawBase is the raw.githubusercontent.com CDN base for the configs directory.
// Both index.json and every walkthrough YAML are served from here.
const rawBase = "https://raw.githubusercontent.com/XotoX1337/GoThrough/main/configstore/configs/"

// Entry describes one walkthrough config listed in the catalog. Path is relative
// to the configs directory (e.g. "gothic2/day1.yaml"); Hash is the sha256-hex of
// the YAML file's bytes, used to detect updates.
type Entry struct {
	Game    string `json:"game"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Chapter int    `json:"chapter"`
	Path    string `json:"path"`
	Hash    string `json:"hash"`
}

// Index is the table of contents (index.json): the catalog of available configs.
type Index struct {
	Schema  int     `json:"schema"`
	Configs []Entry `json:"configs"`
}

// sortEntries orders the catalog by game then chapter so the picker is stable.
func sortEntries(es []Entry) {
	sort.Slice(es, func(i, j int) bool {
		if es[i].Game != es[j].Game {
			return es[i].Game < es[j].Game
		}
		if es[i].Chapter != es[j].Chapter {
			return es[i].Chapter < es[j].Chapter
		}
		return es[i].Path < es[j].Path
	})
}

// ListEmbedded returns the catalog from the index.json embedded in the binary —
// the offline fallback when the remote index can't be fetched.
func ListEmbedded() []Entry {
	var idx Index
	if err := json.Unmarshal(embeddedIndex, &idx); err != nil {
		return nil
	}
	sortEntries(idx.Configs)
	return idx.Configs
}

// ListRemote fetches index.json from the CDN and returns its catalog. Any
// network/timeout/status error is returned so the caller can fall back to
// ListEmbedded.
func ListRemote(ctx context.Context) ([]Entry, error) {
	data, err := fetch(ctx, rawBase+"index.json")
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing remote index: %w", err)
	}
	sortEntries(idx.Configs)
	return idx.Configs, nil
}

// FetchConfig downloads one walkthrough YAML by its catalog-relative path.
func FetchConfig(ctx context.Context, relpath string) ([]byte, error) {
	return fetch(ctx, rawBase+path.Clean(relpath))
}

func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GoThrough")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// --- on-disk cache -------------------------------------------------------

const cacheIndexName = "index.json"

// CacheDir is os.UserConfigDir()/GoThrough/configs — where downloaded configs
// live alongside progress.json and settings.json.
func CacheDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config dir: %w", err)
	}
	return filepath.Join(dir, "GoThrough", "configs"), nil
}

// CachePath returns the on-disk path for a catalog-relative config path.
func CachePath(relpath string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filepath.FromSlash(path.Clean(relpath))), nil
}

// ReadCached returns the cached YAML bytes for a catalog-relative path. A missing
// file yields an error (the caller surfaces "not downloaded / offline").
func ReadCached(relpath string) ([]byte, error) {
	p, err := CachePath(relpath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

// readCacheIndex reads the local cache index — the mirror of downloaded entries
// and their hashes. A missing or unreadable index is an empty catalog, not an
// error.
func readCacheIndex() Index {
	dir, err := CacheDir()
	if err != nil {
		return Index{Schema: 1}
	}
	data, err := os.ReadFile(filepath.Join(dir, cacheIndexName))
	if err != nil {
		return Index{Schema: 1}
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return Index{Schema: 1}
	}
	if idx.Schema == 0 {
		idx.Schema = 1
	}
	return idx
}

func writeCacheIndex(idx Index) error {
	dir, err := CacheDir()
	if err != nil {
		return err
	}
	sortEntries(idx.Configs)
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(dir, cacheIndexName), data)
}

// atomicWrite writes data to dst via a temp file + rename so a crash mid-write
// can't leave a half-written file (mirrors the progress/settings stores).
func atomicWrite(dst string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dst)
}

// ClearCache removes the entire on-disk config cache (every downloaded YAML and
// the local cache index). A missing cache directory is not an error. After this
// the picker treats every game as not-yet-downloaded, so games are re-fetched
// from the catalog on demand.
func ClearCache() error {
	dir, err := CacheDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("clearing config cache: %w", err)
	}
	return nil
}

// DownloadGame downloads every chapter of game (from the given catalog) into the
// cache and records each chapter's hash in the local cache index, so the chapter
// list loads from disk — instant, offline-capable, and including `next:`
// hand-offs across files. Already-cached chapters are refreshed in place.
//
// It returns nil when every chapter of the game is present on disk afterward
// (so an offline click on an already fully-cached game still succeeds); it
// returns the first fetch error only when some chapter is neither freshly
// downloaded nor already cached.
func DownloadGame(ctx context.Context, catalog []Entry, game string) error {
	idx := readCacheIndex()
	pos := map[string]int{}
	for i, e := range idx.Configs {
		pos[e.Path] = i
	}

	wanted := 0
	var fetchErr error
	for _, e := range catalog {
		if e.Game != game {
			continue
		}
		wanted++
		data, err := FetchConfig(ctx, e.Path)
		if err != nil {
			if fetchErr == nil {
				fetchErr = err
			}
			continue
		}
		p, err := CachePath(e.Path)
		if err != nil {
			if fetchErr == nil {
				fetchErr = err
			}
			continue
		}
		if err := atomicWrite(p, data); err != nil {
			if fetchErr == nil {
				fetchErr = err
			}
			continue
		}
		rec := e
		rec.Hash = hashBytes(data)
		if i, ok := pos[e.Path]; ok {
			idx.Configs[i] = rec
		} else {
			idx.Configs = append(idx.Configs, rec)
			pos[e.Path] = len(idx.Configs) - 1
		}
	}
	_ = writeCacheIndex(idx)

	if wanted == 0 {
		return fmt.Errorf("no chapters for game %q", game)
	}
	// Success only if every wanted chapter exists on disk now (already-cached
	// chapters count, so offline-with-cache works); otherwise report the error.
	for _, e := range catalog {
		if e.Game != game {
			continue
		}
		p, err := CachePath(e.Path)
		if err != nil {
			return fetchErr
		}
		if _, err := os.Stat(p); err != nil {
			return fetchErr
		}
	}
	return nil
}

// RefreshUpdates re-downloads cached chapters whose remote hash changed and pulls
// newly-added chapters for games that are already cached, given the freshly
// fetched remote catalog. It is meant to run in the background at startup and
// returns the number of chapters updated. Games not present in the cache are
// left untouched (the user downloads those on demand by picking the game).
func RefreshUpdates(ctx context.Context, remote []Entry) int {
	idx := readCacheIndex()
	if len(idx.Configs) == 0 {
		return 0
	}
	cachedGames := map[string]bool{}
	pos := map[string]int{}
	for i, e := range idx.Configs {
		cachedGames[e.Game] = true
		pos[e.Path] = i
	}

	updated := 0
	for _, e := range remote {
		if !cachedGames[e.Game] {
			continue
		}
		if i, ok := pos[e.Path]; ok && idx.Configs[i].Hash == e.Hash {
			continue // up to date
		}
		data, err := FetchConfig(ctx, e.Path)
		if err != nil {
			continue
		}
		p, err := CachePath(e.Path)
		if err != nil {
			continue
		}
		if err := atomicWrite(p, data); err != nil {
			continue
		}
		rec := e
		rec.Hash = hashBytes(data)
		if i, ok := pos[e.Path]; ok {
			idx.Configs[i] = rec
		} else {
			idx.Configs = append(idx.Configs, rec)
			pos[e.Path] = len(idx.Configs) - 1
		}
		updated++
	}
	if updated > 0 {
		_ = writeCacheIndex(idx)
	}
	return updated
}
