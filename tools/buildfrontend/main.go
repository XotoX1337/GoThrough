// Command buildfrontend transpiles the overlay HUD's TypeScript source
// (overlay/frontend/src/app.ts) into the committed, //go:embed'd app.js via the
// esbuild Go API (no npm — see package frontendbuild and CLAUDE.md).
//
// It is wired as a //go:generate directive in overlay/overlay.go; run
// `go generate ./...` before `wails build -s` after editing app.ts. app.js is
// committed, so a plain build needs no generate step; CI runs generate to catch
// drift.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/XotoX1337/GoThrough/tools/frontendbuild"
)

func main() {
	out, err := frontendbuild.Build()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	fmt.Printf("buildfrontend: wrote %s\n", out)
}
