// Package webapp embeds the built frontend (web/, built via `pnpm build`
// with its Vite output directory pointed at internal/webapp/dist) so the
// server binary can serve it directly with no separate static file
// container.
package webapp

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var distFS embed.FS

// Handler serves the embedded frontend build. Until `pnpm build` has run at
// least once, dist/ only contains a placeholder index.html.
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}
	return http.FileServer(http.FS(sub)), nil
}
