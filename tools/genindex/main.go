// Command genindex regenerates configstore/configs/index.json — the catalog the
// binary embeds and the app refreshes from the CDN at runtime.
//
// It walks every walkthrough YAML under configstore/configs, parses each via
// config.LoadBytes (so a broken file fails the build instead of silently
// dropping out of the catalog), and writes an index.json with one entry per
// config: game/title/author/chapter metadata, the catalog-relative path, and a
// sha256-hex hash of the file's bytes used for update detection. Entries are
// stably sorted by game then chapter.
//
// This is run by the index.yml GitHub workflow on any push that touches a
// config YAML — it never needs to run locally.
//
//	go run ./tools/genindex
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/XotoX1337/GoThrough/config"
)

const configsDir = "configstore/configs"

type entry struct {
	Game    string `json:"game"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Chapter int    `json:"chapter"`
	Path    string `json:"path"`
	Hash    string `json:"hash"`
}

type index struct {
	Schema  int     `json:"schema"`
	Configs []entry `json:"configs"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "genindex:", err)
		os.Exit(1)
	}
}

func run() error {
	var entries []entry
	err := filepath.WalkDir(configsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(p, ".yaml") && !strings.HasSuffix(p, ".yml") {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		wt, err := config.LoadBytes(data)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		rel, err := filepath.Rel(configsDir, p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		entries = append(entries, entry{
			Game:    wt.Game,
			Title:   wt.Title,
			Author:  wt.Author,
			Chapter: wt.Chapter,
			Path:    filepath.ToSlash(rel),
			Hash:    hex.EncodeToString(sum[:]),
		})
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Game != entries[j].Game {
			return entries[i].Game < entries[j].Game
		}
		if entries[i].Chapter != entries[j].Chapter {
			return entries[i].Chapter < entries[j].Chapter
		}
		return entries[i].Path < entries[j].Path
	})

	out, err := json.MarshalIndent(index{Schema: 1, Configs: entries}, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	dst := filepath.Join(configsDir, "index.json")
	if err := os.WriteFile(dst, out, 0o644); err != nil {
		return err
	}
	fmt.Printf("genindex: wrote %s (%d configs)\n", dst, len(entries))
	return nil
}
