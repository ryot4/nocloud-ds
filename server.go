package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("failed to parse remote address %s: %v", r.RemoteAddr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		clientIP = strings.Split(xff, " ")[0]
	}

	if r.Method == http.MethodGet {
		handler.serveInstanceData(sw, r, clientIP)
	} else {
		sw.WriteHeader(http.StatusMethodNotAllowed)
	}
	log.Printf("%s \"%s %s %s\" %d \"%s\" %s",
		r.RemoteAddr,
		r.Method,
		r.URL.String(),
		r.Proto,
		sw.statusCode,
		r.UserAgent(),
		clientIP)
}

func (handler *DataSourceHandler) serveInstanceData(w http.ResponseWriter, r *http.Request, clientIP string) {
	switch r.URL.Path {
	case "/meta-data":
		handler.serveFile(w, r, clientIP, "meta-data")
	case "/user-data":
		handler.serveFile(w, r, clientIP, "user-data")
	case "/vendor-data":
		handler.serveFile(w, r, clientIP, "vendor-data")
	case "/network-config":
		handler.serveFile(w, r, clientIP, "network-config")
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (handler *DataSourceHandler) serveFile(w http.ResponseWriter, r *http.Request, clientIP, name string) {
	instanceDir := filepath.Join(handler.dataSourceDir, clientIP)
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
