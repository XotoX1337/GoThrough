// Package frontendbuild transpiles the overlay HUD's frontend sources into the
// committed, //go:embed'd bundles using the esbuild Go API: the TypeScript
// source (overlay/frontend/src/app.ts → app.js) and the CSS source
// (overlay/frontend/src/app.css → app.css). esbuild is written in Go and pulled
// via go.mod (checksum-verified, no npm / no node_modules / no lifecycle
// scripts), keeping the project a single Go toolchain — see CLAUDE.md. esbuild
// only strips TS types here (no type-checking), which is sufficient for the
// build. Both bundles are minified — the committed app.js/app.css are shipped
// artifacts (smaller for the WebView to load), not meant to be read as diffs;
// the readable source stays in src/app.ts and src/app.css.
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

// CSSEntryPoint is the CSS source, CSSOutFile the generated CSS bundle.
func CSSEntryPoint() string {
	return filepath.Join(repoRoot(), "overlay", "frontend", "src", "app.css")
}

func CSSOutFile() string {
	return filepath.Join(repoRoot(), "overlay", "frontend", "app.css")
}

// Build transpiles app.ts → app.js and app.css → app.css. It returns an error
// describing the first esbuild diagnostic on failure; on success it returns the
// JS output path (the CSS path is CSSOutFile()).
func Build() (string, error) {
	out := OutFile()
	js := api.Build(api.BuildOptions{
		EntryPoints: []string{EntryPoint()},
		Outfile:     out,
		Bundle:      true,
		Write:       true,
		Format:      api.FormatIIFE, // classic <script>, no ESM needed
		Target:      api.ES2020,     // WebView2 / Chromium-Edge → modern JS is fine
		Charset:     api.CharsetUTF8,
		// Minified shipped artifact; window.go/window.runtime are external
		// globals, so MinifyIdentifiers never touches them.
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		LogLevel:          api.LogLevelSilent,
	})
	if len(js.Errors) > 0 {
		return "", esbuildError(js.Errors)
	}

	css := api.Build(api.BuildOptions{
		EntryPoints: []string{CSSEntryPoint()},
		Outfile:     CSSOutFile(),
		Bundle:      true,        // resolves @import if the CSS is ever split
		Write:       true,
		Loader:      map[string]api.Loader{".css": api.LoaderCSS},
		Target:      api.ES2020,  // lower CSS nesting etc. for WebView2
		Charset:     api.CharsetUTF8,
		// Minified shipped artifact (see the JS build above).
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		LogLevel:          api.LogLevelSilent,
	})
	if len(css.Errors) > 0 {
		return "", esbuildError(css.Errors)
	}
	return out, nil
}

func esbuildError(errs []api.Message) error {
	msgs := api.FormatMessages(errs, api.FormatMessagesOptions{Color: false, Kind: api.ErrorMessage})
	return fmt.Errorf("esbuild: %d error(s):\n%s", len(errs), join(msgs))
}

func join(msgs []string) string {
	s := ""
	for _, m := range msgs {
		s += m
	}
	return s
}
