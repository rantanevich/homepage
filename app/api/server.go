package api

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/watcher"
	"github.com/rantanevich/homepage/app/web"
)

type Server struct {
	log *slog.Logger

	indexPage  []byte
	httpServer *http.Server
	templates  *template.Template
}

func New(log *slog.Logger) *Server {
	return &Server{
		log:       log.With(slog.String("component", "api")),
		templates: template.Must(template.ParseFS(web.WebFS, "templates/*")),
	}
}

func (s *Server) Run(port int, iconsDir string) {
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           s.setupRouter(iconsDir),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	s.log.Info("started", slog.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		s.log.Error("terminated", slog.String("error", err.Error()))
		return
	}
	s.log.Info("stopped", slog.String("addr", s.httpServer.Addr))
}

func (s *Server) Shutdown(ctx context.Context) {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("failed to shutdown", slog.String("error", err.Error()))
	}
}

func (s *Server) UpdateIndexPage(conf dynamic.Config, title, logo string) {
	tmplData := struct {
		Title  string
		Logo   string
		Groups dynamic.Config
	}{
		Title:  title,
		Logo:   watcher.ResolveIcon(logo),
		Groups: conf,
	}

	page := bytes.NewBuffer(nil)
	err := s.templates.ExecuteTemplate(page, "index.tmpl", tmplData)
	if err != nil {
		s.log.Error("failed to render index.html", slog.String("error", err.Error()))
		return
	}
	s.indexPage = page.Bytes()
}
