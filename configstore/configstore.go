// Package configstore provides the embedded set of community walkthrough
// configs that ship with the GoThrough binary, and a background probe to
// detect new configs in the public GitHub repo.
package configstore

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strings"

	"github.com/XotoX1337/GoThrough/config"
)

//go:embed configs
var embedded embed.FS

// Entry describes a walkthrough config — either bundled in the binary or
// discovered in the remote repo.
type Entry struct {
	Game     string `json:"game"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Chapter  int    `json:"chapter"`
	Path     string `json:"path"`
	Embedded bool   `json:"embedded"`
}

// List returns all configs bundled in the binary, sorted by game then chapter.
func List() []Entry {
	var entries []Entry
	_ = fs.WalkDir(embedded, "configs", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, err := embedded.ReadFile(path)
		if err != nil {
			return nil
		}
		wt, err := config.LoadBytes(data)
		if err != nil {
			return nil
		}
		entries = append(entries, Entry{
			Game:     wt.Game,
			Title:    wt.Title,
			Author:   wt.Author,
			Chapter:  wt.Chapter,
			Path:     path,
			Embedded: true,
		})
		return nil
	})
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Game != entries[j].Game {
			return entries[i].Game < entries[j].Game
		}
		return entries[i].Chapter < entries[j].Chapter
	})
	return entries
}

// Open returns the raw YAML bytes of an embedded config at path.
func Open(path string) ([]byte, error) {
	return embedded.ReadFile(path)
}

type githubTreeNode struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type githubTree struct {
	Tree []githubTreeNode `json:"tree"`
}

// FetchNewRemote queries the GitHub API and returns entries for configs that
// exist in the repo but are NOT in the embedded set. Returns nil, nil when
// the repo has no configs newer than the binary. Network errors are returned
// so callers can log them as informational (the binary is fully usable offline).
func FetchNewRemote(ctx context.Context) ([]Entry, error) {
	const url = "https://api.github.com/repos/XotoX1337/GoThrough/git/trees/main?recursive=1"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "GoThrough")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	var tree githubTree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, err
	}

	known := make(map[string]bool)
	for _, e := range List() {
		known[e.Path] = true
	}

	var novel []Entry
	for _, node := range tree.Tree {
		if node.Type != "blob" ||
			!strings.HasPrefix(node.Path, "configs/") ||
			!strings.HasSuffix(node.Path, ".yaml") {
			continue
		}
		if !known[node.Path] {
			novel = append(novel, Entry{Path: node.Path})
		}
	}
	return novel, nil
}
