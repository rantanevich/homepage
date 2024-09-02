package main

import (
	"bytes"
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"
)

var (
	//go:embed web/*
	webFS embed.FS

	indexBody []byte
)

func main() {
	conf, err := LoadConfig()
	fatal(err)

	body, err := renderIndexPage(conf)
	fatal(err)
	indexBody = body

	router, err := setupRouter(conf.IconsDir)
	fatal(err)

	srv := &http.Server{
		Addr:              ":3000",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("[INFO] http server started on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("[ERROR] %v", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("[INFO] http server is stopping...")
	fatal(srv.Shutdown(ctx))
}

func setupRouter(iconsDir string) (*http.ServeMux, error) {
	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler())
	mux.Handle(staticPattern, http.StripPrefix(staticPattern, http.FileServer(http.FS(staticFS))))
	mux.Handle(iconsPattern, http.StripPrefix(iconsPattern, http.FileServer(http.Dir(iconsDir))))

	return mux, nil
}

func indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(indexBody)
	}
}

func renderIndexPage(conf *Config) ([]byte, error) {
	tmpl, err := template.ParseFS(webFS, "web/templates/index.tmpl")
	if err != nil {
		return nil, err
	}

	page := bytes.NewBuffer(nil)
	if err := tmpl.Execute(page, conf); err != nil {
		return nil, err
	}
	return page.Bytes(), nil
}
