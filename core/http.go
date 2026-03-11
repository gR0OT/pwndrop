package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kgretzky/pwndrop/log"
)

const BLACKLIST_JAIL_TIME_SECS = 10 * 60
const BLACKLIST_HITS_LIMIT = 10

type BlacklistItem struct {
	hits     int
	last_hit time.Time
}

type Http struct {
	srv *Server
}

func NewHttp(srv *Server) (*Http, error) {
	s := &Http{
		srv: srv,
	}
	return s, nil
}

func (s *Http) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	data_dir := Cfg.GetDataDir()

	from_ip := request.RemoteAddr
	if strings.Contains(from_ip, ":") {
		from_ip = strings.Split(from_ip, ":")[0]
	}

	if request.Method == "GET" {
		// file, status, err := s.srv.GetFileHTTP(request.URL.Path)
		file, status, err := s.srv.GetFileHTTP(request)

		if err != nil {
			log.Error("http: get: %s: %s (%s)", request.URL.Path, err, from_ip)
			err := s.killConnection(response, status)
			if err != nil {
				log.Error("http: %s (%s)", err, from_ip)
			}
			return
		}

		if file.RedirectPath != "" && file.RedirectPath != request.URL.Path && !file.IsPaused {
			log.Error("http: get: %s: redirecting to '%s' (%s)", request.URL.Path, file.RedirectPath, from_ip)
			http.Redirect(response, request, file.RedirectPath, http.StatusFound)
		} else {
			mime_type := file.MimeType
			if file.IsPaused {
				mime_type = file.SubMimeType
			}

			if file.GetParamEnabled {
				mime_type = file.MimeType
			}

			fpath := filepath.Join(data_dir, "files", file.Filename)
			fo, err := os.Open(fpath)
			//data, err := ioutil.ReadFile(fpath)
			if err != nil {
				log.Error("http: file: %s: %s (%s)", file.Filename, err, from_ip)
				return
			}
			defer fo.Close()

			response.Header().Set("Content-Type", mime_type)
			response.WriteHeader(200)
			io.Copy(response, fo)
		}
		return
	}
	err := s.killConnection(response, 404)
	if err != nil {
		log.Error("http: %s (%s)", err, from_ip)
	}
}

func (s *Http) killConnection(w http.ResponseWriter, status int) error {
	if status > 0 {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(status)
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("connection hijacking not supported")
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
