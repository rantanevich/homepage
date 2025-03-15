package api

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/rantanevich/homepage/app/web"
)

func (s *Server) setupRouter(iconsDir string) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/", s.indexHandler())

	staticFS, err := fs.Sub(web.WebFS, "static")
	if err == nil {
		router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	} else {
		s.log.Error(
			"cannot start fileserver",
			slog.String("pattern", "/static/"),
			slog.String("error", err.Error()),
		)
	}

	if err := os.MkdirAll(iconsDir, 0o755); err == nil {
		router.Handle("/icons/", http.StripPrefix("/icons/", http.FileServer(http.Dir(iconsDir))))
	} else {
		s.log.Error(
			"cannot start fileserver",
			slog.String("directory", iconsDir),
			slog.String("pattern", "/icons/"),
			slog.String("error", err.Error()),
		)
	}

	return router
}

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(s.indexPage)
	}
}
