// Package frontendbuild transpiles the overlay HUD's TypeScript source
// (overlay/frontend/src/app.ts) into the committed, //go:embed'd app.js using
// the esbuild Go API. esbuild is written in Go and pulled via go.mod (checksum-
// verified, no npm / no node_modules / no lifecycle scripts), keeping the
// project a single Go toolchain — see CLAUDE.md. esbuild only strips types here
// (no type-checking), which is sufficient for the build.
//
// Both the buildfrontend command (run via `go generate` / `go run`) and devui
// call Build, so the transpile path can't drift between them.
package frontendbuild

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/evanw/esbuild/pkg/api"
)

// repoRoot resolves the repository root from this source file's location, so the
// build works regardless of the caller's working directory (`go generate` runs
// with CWD set to the overlay package dir; `go run ./tools/...` runs from the
// repo root). This file lives at <root>/tools/frontendbuild/frontendbuild.go.
func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// EntryPoint is the TypeScript source, OutFile the generated JS bundle.
func EntryPoint() string {
	return filepath.Join(repoRoot(), "overlay", "frontend", "src", "app.ts")
}

func OutFile() string {
	return filepath.Join(repoRoot(), "overlay", "frontend", "app.js")
}

// Build transpiles app.ts to app.js. It returns an error describing the first
// esbuild diagnostic on failure; the written path is returned on success.
func Build() (string, error) {
	out := OutFile()
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{EntryPoint()},
		Outfile:     out,
		Bundle:      true,
		Write:       true,
		Format:      api.FormatIIFE, // classic <script>, no ESM needed
		Target:      api.ES2020,     // WebView2 / Chromium-Edge → modern JS is fine
		Charset:     api.CharsetUTF8,
		LogLevel:    api.LogLevelSilent,
	})
	if len(result.Errors) > 0 {
		msgs := api.FormatMessages(result.Errors, api.FormatMessagesOptions{Color: false, Kind: api.ErrorMessage})
		return "", fmt.Errorf("esbuild: %d error(s):\n%s", len(result.Errors), join(msgs))
	}
	return out, nil
}

func join(msgs []string) string {
	s := ""
	for _, m := range msgs {
		s += m
	}
	return s
}
