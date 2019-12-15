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
	r := mux.NewRouter()

	r.HandleFunc("/downloads", s.downloadsCreate).Methods("POST")
	r.HandleFunc("/downloads/{id}", s.downloadsShow).Methods("GET")
	r.HandleFunc("/", s.index).Methods("GET")

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, os.Kill)
	terminateCh := make(chan error)
	go s.handleShutdown(signalCh, terminateCh, cleanUpFunc)

	manners.ListenAndServe(fmt.Sprintf(":%d", s.port), r)

	// Do not stop program until receive message on terminate channel.
	shutdownErr := <-terminateCh
	return shutdownErr
}

func (s *Server) handleShutdown(signalCh <-chan os.Signal, terminateCh chan<- error, cleanUpFunc func() error) {
	<-signalCh
	s.logger.Info("Handling shutdown")
	manners.Close()
	terminateCh <- cleanUpFunc()
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/index.html"))
	t.Execute(w, nil)
}

func (s *Server) downloadsCreate(w http.ResponseWriter, r *http.Request) {
	// TODO: Decide how I want to handle http errors.
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse form: %s", err), http.StatusInternalServerError)
		return
	}

	remotePath := r.FormValue("url")

	localFilePath, err := s.contentDownloader.DownloadContent(remotePath, &DownloadOptions{audioOnly: true})
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to download content: %s", err), http.StatusInternalServerError)
		return
	}

	publicURL, err := s.contentUploader.UploadContentPublicly(localFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to upload content: %s", err), http.StatusInternalServerError)
		return
	}

	// So can we redirect via an id... Do we need to make accessing this url
	// cache threadsafe?
	downloadId := generateRandomString(defaultRandomStringLength)
	s.publicDownloadURLCache[downloadId] = publicURL

	http.Redirect(w, r, fmt.Sprintf("/downloads/%s", downloadId), http.StatusSeeOther)
}

// TODO: Naming convention for objects containing template vars...
type downloadShowPage struct {
	PublicDownloadURL string
}

func (s *Server) downloadsShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	publicDownloadURL, found := s.publicDownloadURLCache[vars["id"]]
	if !found {
		http.NotFound(w, r)
	}

	p := &downloadShowPage{
		PublicDownloadURL: publicDownloadURL,
	}

	t := template.Must(template.ParseFiles("templates/download.html"))
	t.Execute(w, p)
}
