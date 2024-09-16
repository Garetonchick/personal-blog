package server

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Garetonchick/personal-blog/internal/articles"
	"github.com/Garetonchick/personal-blog/internal/utils"
)

type Server struct {
	mux             *http.ServeMux
	articlesManager *articles.Manager
	serveStatic     http.Handler
}

type templateHandler func(rw http.ResponseWriter, r *http.Request) (files []string, data any, err error)
type errorHandler func(rw http.ResponseWriter, r *http.Request) error

func New() (*Server, error) {
	var srv Server
	srv.serveStatic = http.FileServer(http.Dir("./static"))

	workdir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	srv.articlesManager, err = articles.NewManager(workdir)
	if err != nil {
		return nil, err
	}

	srv.mux = http.NewServeMux()
	srv.registerHandlers()

	return &srv, nil
}

func (s *Server) registerHandlers() {
	apply := func(th templateHandler) http.Handler {
		return errorMiddleware(templateMiddleware(th))
	}

	s.mux.Handle("GET /static/", http.StripPrefix("/static/", s.serveStatic))

	s.mux.Handle("GET /home", apply(s.ServeHomePage))
	s.mux.Handle("GET /articles/{id}", apply(s.ServeArticlePage))
	s.mux.Handle("GET /articles/edit/{id}", apply(s.ServeArticlePage))
}

func (s *Server) ServeArticleEditPage(rw http.ResponseWriter, r *http.Request) {

}

func errorMiddleware(h errorHandler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		err := h(rw, r)
		if err != nil {
			log.Printf("ERROR: %q", err)
			rw.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func templateMiddleware(h templateHandler) errorHandler {
	return errorHandler(func(rw http.ResponseWriter, r *http.Request) error {
		files, data, err := h(rw, r)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return nil
		}
		for i := range files {
			files[i] = filepath.Join("templates", files[i])
		}
		layoutFile := filepath.Join("templates", "layout.html")
		files = append([]string{layoutFile}, files...)

		tpl, err := template.ParseFiles(files...)
		if err != nil {
			return err
		}

		return tpl.ExecuteTemplate(rw, "layout", data)
	})
}

func (s *Server) ServeHomePage(rw http.ResponseWriter, r *http.Request) ([]string, any, error) {
	ids := s.articlesManager.List()
	ids = ids[:min(10, len(ids))]
	var ars []*articles.Article

	for _, id := range ids {
		a := s.articlesManager.Load(id)
		if a == nil {
			continue
		}
		ars = append(ars, a)
	}

	return []string{"home.html"}, ars, nil
}

func (s *Server) ServeArticlePage(rw http.ResponseWriter, r *http.Request) ([]string, any, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid id: %w", err)
	}

	a := s.articlesManager.Load(id)
	if a == nil {
		return nil, nil, errors.New("article not found")
	}

	htmlArticle := struct {
		HTMLContent template.HTML
		A           *articles.Article
	}{
		HTMLContent: utils.MD2SafeHTML(a.Content),
		A:           a,
	}

	return []string{"article.html"}, htmlArticle, nil
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}
