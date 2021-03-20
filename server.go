package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	// Remember the status code for later logging.
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

type DataSourceHandler struct {
	dataSourceDir string
}

func (handler *DataSourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sw := &statusResponseWriter{ResponseWriter: w}

	if r.Method == http.MethodGet {
		handler.serveInstanceData(sw, r)
	} else {
		sw.WriteHeader(http.StatusMethodNotAllowed)
	}
	log.Printf("%s \"%s %s %s\" %d \"%s\"",
		r.RemoteAddr,
		r.Method,
		r.URL.String(),
		r.Proto,
		sw.statusCode,
		r.UserAgent())
}

func (handler *DataSourceHandler) serveInstanceData(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/meta-data":
		handler.serveFile(w, r, "meta-data")
	case "/user-data":
		handler.serveFile(w, r, "user-data")
	case "/vendor-data":
		handler.serveFile(w, r, "vendor-data")
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (handler *DataSourceHandler) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("failed to parse remote address %s: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	instanceDir := filepath.Join(handler.dataSourceDir, net.ParseIP(host).String())
	fi, err := os.Stat(instanceDir)
	if os.IsNotExist(err) {
		// Try to serve the default file
		http.ServeFile(w, r, filepath.Join(handler.dataSourceDir, name))
		return
	} else if err != nil {
		log.Printf("failed to stat %s: %v", instanceDir, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !fi.IsDir() {
		log.Printf("%s exists but is not a directory", instanceDir)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filepath.Join(instanceDir, name))
}

func NewServer(listenAddress, dataSourceDir string) *http.Server {
	return &http.Server{
		Addr:    listenAddress,
		Handler: &DataSourceHandler{dataSourceDir},
	}
}
