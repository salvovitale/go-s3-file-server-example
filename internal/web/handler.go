package web

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/salvovitale/go-s3-file-server-example/internal/store"
)

func NewHandler(s store.Store, csrfKey []byte) *Handler {
	h := &Handler{
		Mux:   chi.NewRouter(),
		store: s,
	}

	fileHandler := FileHandler{store: s}

	// add logger middleware
	h.Use(middleware.Logger)

	// add csrf protection middleware
	h.Use(csrf.Protect(csrfKey, csrf.Secure(false))) // set security to false for development otherwise the cookie will only be sent over https

	// add custom middleware to retrieve the user from the session and add it to the request context
	// h.Use(h.withUser)

	// homepage
	h.Get("/", h.homeView())

	// sub paths
	h.Route("/files", func(r chi.Router) {
		// 	r.Get("/", fileHandler.listFilesView())
		r.Get("/upload", fileHandler.uploadView())
		// 	r.Get("/{id}", fileHandler.view())
		r.Post("/upload", fileHandler.upload())
		// 	r.Post("/{id}/delete", fileHandler.delete())
	})

	return h
}

type Handler struct {
	*chi.Mux //embedded structure
	store    store.Store
}

func (h *Handler) homeView() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}
