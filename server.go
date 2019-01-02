package main

import (
	"context"
	"fallout76_ss/api_calls"
	"fallout76_ss/api_db"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type key int

const (
	requestIDKey key = 0
)

var (
	listenAddr string
)


// Result of our database return for the homepage
// Note: we do this since querying the database has become slow and I am too lazy to optimize it
// Note: thus we just refresh the data every few minutes
var homeData api_db.HomepageData
var homeTicker time.Ticker


// Main function
// This will start the server and also start the ticker
// This ticker will poll the API, while the webserver will serve traffic
// Based off of: https://github.com/Xeoncross/vanilla-go-server/blob/master/server.go
func main() {

	// Parse the listen port
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "server listen address")
	flag.Parse()

	// Start the API ticker
	api_calls.SetPollingFrequency(60)
	api_calls.StartAPITicker()

	// Make our logger
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Make our homepage data poller (updates every two minute)
	homeTicker := time.NewTicker(120*time.Second)
	go func() {
		for ; true; <-homeTicker.C {
			// Open our database connection
			db := api_db.OpenDatabase()
			defer db.Close()
			// Get the data from our database object
			homeData = api_db.GetHomepageData(db)
		}
	}()

	// Now create the server
	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(routes())),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Server the server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	// end of program stop ticker
	homeTicker.Stop()
	api_calls.StopAPITicker()
	logger.Println("Server and Ticker stopped")

}

// Setup all your routes
func routes() *http.ServeMux {

	// Create a empty http router
	router := http.NewServeMux()

	// Load the main page template
	tmpl := template.Must(template.ParseFiles("web_tpl/main.html"))

	// The homepage!
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Return 404 if not homepage
		if r.URL.Path != "/" {
			http.Redirect(w, r, "/", 301)
			return
		}
		// Open our database connection
		db := api_db.OpenDatabase()
		defer db.Close()
		// Get the data from our database object (just update the current status)
		status, strtime := api_db.GetCurrentStatus(db)
		homeData.StatusOnline = (status == 1)
		homeData.StatusOffline = (status == 2)
		homeData.StatusUnknown = (status != 1 && status != 2)
		homeData.TimeAgoString = strtime
		// Render the template
		tmpl.Execute(w, homeData)
	})

	// Here we convert all the "static" paths to where our static js/css files are stored
	// We use a custom filesystem to remove the default directory listing
	// https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
	fs := http.FileServer(neuteredFileSystem{http.Dir("web_static/")})
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	// Return the router with the routes...
	return router
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := nfs.fs.Open(index); err != nil {
			return nil, err
		}
	}
	return f, nil
}
