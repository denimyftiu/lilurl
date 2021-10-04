package shortner

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	l  log.Logger
	sh Shortner
}

type ShortenRequest struct {
	Url string `json:"url"`
}

type ShortenResponse struct {
	Token string `json:"token"`
}

func NewServer(sh Shortner) *Server {
	return &Server{sh: sh}
}

func (s Server) Shorten(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	token, err := s.sh.Shorten(r.Context(), req.Url)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(token)
	return
}

func (s Server) Expand(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	p := mux.Vars(r)
	token, ok := p["token"]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	url, err := s.sh.Expand(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			rw.WriteHeader(http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(rw, r, url, http.StatusMovedPermanently)
	return
}

// We may need to pass the mux to this. For flexibility.
func (s Server) Install() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", http.HandlerFunc(s.Shorten)).Methods("POST")
	r.Handle("/{token}", http.HandlerFunc(s.Expand)).Methods("GET")
	return r
}
