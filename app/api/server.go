package api

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/rantanevich/homepage/app/config/dynamic"
	"github.com/rantanevich/homepage/app/watcher"
	"github.com/rantanevich/homepage/app/web"
)

type Server struct {
	indexPage  []byte
	httpServer *http.Server
	templates  *template.Template
}

func New() *Server {
	return &Server{
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

	log.Printf("[INFO] http server started on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("[ERROR] http server terminated: %v", err)
		return
	}
	log.Printf("[INFO] http server stopped gracefully")
}

func (s *Server) Shutdown(ctx context.Context) {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] failed to shutdown http server: %v", err)
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
		log.Printf("[ERROR] failed to render index.html: %v", err)
		return
	}
	s.indexPage = page.Bytes()
}
