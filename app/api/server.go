package api

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/rantanevich/homepage/app/config"
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

func getAllFilenames(fs *embed.FS, path string) (out []string, err error) {
	if len(path) == 0 {
		path = "."
	}
	entries, err := fs.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		fp := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			res, err := getAllFilenames(fs, fp)
			if err != nil {
				return nil, err
			}
			out = append(out, res...)
			continue
		}
		out = append(out, fp)
	}
	return
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

func (s *Server) RenderIndexPage(conf *config.Config) error {
	page := bytes.NewBuffer(nil)
	if err := s.templates.ExecuteTemplate(page, "index.tmpl", conf); err != nil {
		return err
	}
	s.indexPage = page.Bytes()
	return nil
}
