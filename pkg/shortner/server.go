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

func NewServer(sh Shortner) *Server {
	return &Server{sh: sh}
}

// We may need to pass the mux to this. For flexibility.
func (s Server) Install() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", http.HandlerFunc(s.Shorten)).Methods("POST")
	r.Handle("/{token}", http.HandlerFunc(s.Expand)).Methods("GET")
	return r
}

type ShortenRequest struct {
	Url string `json:"url"`
}

type ShortenResponse struct {
	Token string `json:"token"`
}

func (s Server) Shorten(rw http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := s.sh.Shorten(r.Context(), req.Url)
	if err != nil {
		if errors.Is(err, ErrorInvalidURL) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	resp := ShortenResponse{Token: token}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(resp)
	return
}

func (s Server) Expand(rw http.ResponseWriter, r *http.Request) {
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
