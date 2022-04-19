package shortner

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	sh *ShortnerService
}

type ServerConfig struct {
	Shortner *ShortnerService
}

func NewServer(scfg *ServerConfig) *Server {
	return &Server{
		sh: scfg.Shortner,
	}
}

// We may need to pass the mux to this. For flexibility.
func (s Server) Install() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", http.HandlerFunc(s.Shorten)).Methods(http.MethodPost)
	r.Handle("/{token}", http.HandlerFunc(s.Expand)).Methods(http.MethodGet)
	return r
}

type ShortenRequest struct {
	Url string `json:"url"`
}

type ShortenResponse struct {
	Token string `json:"token"`
}

func (s *Server) Shorten(rw http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := s.sh.Shorten(r.Context(), req.Url)
	if err != nil {
		if errors.Is(err, ErrorInvalidURL) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	resp := ShortenResponse{Token: token}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(resp)
	return
}

func (s *Server) Expand(rw http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	token, ok := p["token"]
	if !ok {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	url, err := s.sh.Expand(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else if errors.Is(err, ErrorInvalidToken) {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		} else {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(rw, r, url, http.StatusMovedPermanently)
	return
}
