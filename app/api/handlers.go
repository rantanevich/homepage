package api

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/rantanevich/homepage/app/web"
)

func (s *Server) setupRouter(iconsDir string) *http.ServeMux {
	router := http.NewServeMux()

	router.Handle("/", s.recoverMiddleware(s.indexHandler()))

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

func (s *Server) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.log.Error("recovered from panic", slog.String("stacktrace", string(debug.Stack())))
				w.Header().Set("Connection", "close")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(s.indexPage)
	}
}
