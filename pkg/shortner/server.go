package shortner

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strings"

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

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if strings.Compare(mediaType, "application/json") != 0 {
		http.Error(rw, fmt.Sprintf("Invalid content type: %s. Only application/json supported", mediaType), http.StatusUnsupportedMediaType)
		log.Printf("Invalid content type: %s. Only application/json supported", mediaType)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		log.Printf("Failed to decode request %s", err)
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

	rw.Header().Set("Cache-control", "no-store")
	http.Redirect(rw, r, url, http.StatusFound)
}
