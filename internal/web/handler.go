package web

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/salvovitale/go-s3-file-server-example/internal/s3"
	"github.com/salvovitale/go-s3-file-server-example/internal/store"
)

func NewHandler(s store.Store, s3Client *s3.S3Client, bucketName string, csrfKey []byte) *Handler {
	h := &Handler{
		Mux:        chi.NewRouter(),
		store:      s,
		s3Client:   s3Client,
		bucketName: bucketName,
	}

	fileHandler := FileHandler{dbHandler: s, s3Handler: s3Client, bucketName: bucketName}

	// add logger middleware
	h.Use(middleware.Logger)

	// add csrf protection middleware
	h.Use(csrf.Protect(csrfKey, csrf.Secure(false))) // set security to false for development otherwise the cookie will only be sent over https

	// homepage
	h.Get("/", h.homeView())

	// sub paths
	h.Route("/files", func(r chi.Router) {
		r.Get("/upload", fileHandler.uploadView())
		r.Post("/upload", fileHandler.upload())
		r.Get("/{id}/delete", fileHandler.delete())
		r.Get("/{id}/download", fileHandler.download())
	})

	return h
}

type Handler struct {
	*chi.Mux   //embedded structure
	store      store.Store
	s3Client   *s3.S3Client
	bucketName string
}

func (h *Handler) homeView() http.HandlerFunc {
	type data struct {
		Files []store.File
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		ff, err := h.store.Files()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data{Files: ff})
	}
}
