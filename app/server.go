package main

import (
	"fmt"
	"github.com/braintree/manners"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Could define Server interface, but not sure there is any benefit...

// Could use Negroni and Mux, but not sure its necessary right now...
type Server struct {
	port              int
	contentDownloader ContentDownloader
	contentUploader   ContentUploader

	// Should be a richer interface... that we can make either consistent
	// storage or ephemeral.
	publicDownloadURLCache map[string]string

	logger logr.Logger
}

func NewServer(port int, contentDownloader ContentDownloader, contentUploader ContentUploader, logger logr.Logger) *Server {
	publicDownloadURLCache := make(map[string]string)

	return &Server{
		port:                   port,
		contentDownloader:      contentDownloader,
		contentUploader:        contentUploader,
		publicDownloadURLCache: publicDownloadURLCache,
		logger:                 logger,
	}
}

func (s *Server) ListenAndServe(cleanUpFunc func() error) error {
	s.logger.V(2).Info("Creating and launching web server")
	r := mux.NewRouter()

	r.HandleFunc("/downloads", s.downloadsCreate).Methods("POST")
	r.HandleFunc("/downloads/{id}", s.downloadsShow).Methods("GET")
	r.HandleFunc("/", s.index).Methods("GET")

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	terminateCh := make(chan error)
	go s.handleShutdown(signalCh, terminateCh, cleanUpFunc)

	manners.ListenAndServe(fmt.Sprintf(":%d", s.port), r)

	// Do not stop program until receive message on terminate channel.
	shutdownErr := <-terminateCh
	return shutdownErr
}

func (s *Server) handleShutdown(signalCh <-chan os.Signal, terminateCh chan<- error, cleanUpFunc func() error) {
	<-signalCh
	s.logger.V(2).Info("Handling shutdown signal to server")
	manners.Close()
	terminateCh <- cleanUpFunc()
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.logger.V(2).Info("Serving request", "endpoint", "GET#index")
	t := template.Must(template.ParseFiles("templates/index.html"))
	t.Execute(w, nil)
}

func (s *Server) downloadsCreate(w http.ResponseWriter, r *http.Request) {
	s.logger.V(2).Info("Serving request", "endpoint", "POST#downloads")

	// TODO: Decide how I want to handle http errors. We should definitely
	// be logging them.
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse form: %s", err), http.StatusInternalServerError)
		return
	}

	remotePath := r.FormValue("url")

	downloadId := generateRandomString(defaultRandomStringLength)

	// TODO: Should update the `publicDownloadURLCache` to be threadsafe by
	// protecting w/ mutex. Will likely require creating a `downloadIDToURL`
	// interface.
	go func() {
		failed := false

		s.logger.V(3).Info("Starting download", "downloadId", downloadId)
		localFilePath, err := s.contentDownloader.DownloadContent(remotePath, &DownloadOptions{audioOnly: true})
		if err != nil {
			s.logger.V(3).Info("Download failed", "downloadId", downloadId)
			failed = true
		}
		s.logger.V(3).Info("Content download completed", "downloadId", downloadId)

		s.logger.V(3).Info("Starting upload", "downloadId", downloadId)
		publicURL, err := s.contentUploader.UploadContentPublicly(localFilePath)
		if err != nil {
			s.logger.V(3).Info("Upload failed", "downloadId", downloadId)
			failed = true
		}
		s.logger.V(3).Info("Content upload completed", "downloadId", downloadId)

		if !failed {
			s.publicDownloadURLCache[downloadId] = publicURL
		} else {
			// The error handling behavior could definitely be more
			// robust.
			failureMessage := "Failed to download video :("
			s.publicDownloadURLCache[downloadId] = failureMessage
		}
	}()

	s.logger.V(3).Info("Redirecting based on cached download id", "downloadId", downloadId)
	http.Redirect(w, r, fmt.Sprintf("/downloads/%s", downloadId), http.StatusSeeOther)
}

// TODO: Naming convention for objects containing template vars...
type downloadShowPage struct {
	PublicDownloadURL string
	DownloadComplete  bool
}

func (s *Server) downloadsShow(w http.ResponseWriter, r *http.Request) {
	s.logger.V(2).Info("Serving request", "endpoint", "GET#downloads/:id")
	vars := mux.Vars(r)

	p := &downloadShowPage{}

	publicDownloadURL, found := s.publicDownloadURLCache[vars["id"]]

	if found {
		p.PublicDownloadURL = publicDownloadURL
		p.DownloadComplete = true
	}

	t := template.Must(template.ParseFiles("templates/download.html"))
	t.Execute(w, p)
}
