package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type ProxyService struct {
	monolithURL          string
	moviesServiceURL     string
	eventsServiceURL     string
	gradualMigration     bool
	moviesMigrationPercent int
	httpClient           *http.Client
}

func NewProxyService() *ProxyService {
	monolithURL := os.Getenv("MONOLITH_URL")
	if monolithURL == "" {
		monolithURL = "http://monolith:8080"
	}

	moviesServiceURL := os.Getenv("MOVIES_SERVICE_URL")
	if moviesServiceURL == "" {
		moviesServiceURL = "http://movies-service:8081"
	}

	eventsServiceURL := os.Getenv("EVENTS_SERVICE_URL")
	if eventsServiceURL == "" {
		eventsServiceURL = "http://events-service:8082"
	}

	gradualMigration := os.Getenv("GRADUAL_MIGRATION") == "true"
	
	moviesMigrationPercent := 50
	if percentStr := os.Getenv("MOVIES_MIGRATION_PERCENT"); percentStr != "" {
		if percent, err := strconv.Atoi(percentStr); err == nil {
			moviesMigrationPercent = percent
		}
	}

	return &ProxyService{
		monolithURL:          monolithURL,
		moviesServiceURL:     moviesServiceURL,
		eventsServiceURL:     eventsServiceURL,
		gradualMigration:     gradualMigration,
		moviesMigrationPercent: moviesMigrationPercent,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ProxyService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Strangler Fig Proxy is healthy"))
}

func (p *ProxyService) shouldRouteToMoviesService() bool {
	if !p.gradualMigration {
		return false
	}
	return rand.Intn(100) < p.moviesMigrationPercent
}

func (p *ProxyService) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid target URL: %v", err), http.StatusInternalServerError)
		return
	}

	target.Path = r.URL.Path
	target.RawQuery = r.URL.RawQuery

	req, err := http.NewRequest(r.Method, target.String(), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error forwarding request: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (p *ProxyService) handleMovies(w http.ResponseWriter, r *http.Request) {
	var targetURL string
	
	if p.shouldRouteToMoviesService() {
		targetURL = p.moviesServiceURL
		log.Printf("Routing /api/movies to movies-service: %s", targetURL)
	} else {
		targetURL = p.monolithURL
		log.Printf("Routing /api/movies to monolith: %s", targetURL)
	}

	p.proxyRequest(w, r, targetURL)
}

func (p *ProxyService) handleEvents(w http.ResponseWriter, r *http.Request) {
	targetURL := p.eventsServiceURL
	log.Printf("Routing %s to events-service: %s", r.URL.Path, targetURL)
	p.proxyRequest(w, r, targetURL)
}

func (p *ProxyService) handleDefault(w http.ResponseWriter, r *http.Request) {
	targetURL := p.monolithURL
	log.Printf("Routing %s to monolith: %s", r.URL.Path, targetURL)
	p.proxyRequest(w, r, targetURL)
}

func main() {
	proxy := NewProxyService()

	http.HandleFunc("/health", proxy.handleHealth)
	http.HandleFunc("/api/movies", proxy.handleMovies)
	http.HandleFunc("/api/movies/", proxy.handleMovies)
	http.HandleFunc("/api/events/", proxy.handleEvents)
	http.HandleFunc("/api/events", proxy.handleEvents)
	http.HandleFunc("/", proxy.handleDefault)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Starting Strangler Fig Proxy on port %s", port)
	log.Printf("Monolith URL: %s", proxy.monolithURL)
	log.Printf("Movies Service URL: %s", proxy.moviesServiceURL)
	log.Printf("Events Service URL: %s", proxy.eventsServiceURL)
	log.Printf("Gradual Migration: %v", proxy.gradualMigration)
	log.Printf("Movies Migration Percent: %d%%", proxy.moviesMigrationPercent)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

