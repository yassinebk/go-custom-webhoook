package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/fatih/color"
)

type RequestData struct {
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	URL     string            `json:"url"`
}

type Server struct {
	mu       sync.Mutex
	requests []RequestData
}

func (s *Server) saveRequest(request RequestData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Append the request to the requests slice
	s.requests = append(s.requests, request)

	// Save all requests to a JSON file
	data, err := json.MarshalIndent(s.requests, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("requests.json", data, 0644); err != nil {
		return err
	}

	return nil
}

func (s *Server) requestHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Create a RequestData object
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0]
	}

	request := RequestData{
		Method:  r.Method,
		Headers: headers,
		Body:    string(body),
		URL:     r.URL.String(),
	}

	// Print the received request in a nice, colored manner
	color.Cyan("Received Request:")
	color.Green("Method: %s", request.Method)
	color.Blue("URL: %s", request.URL)
	color.Magenta("Headers: %v", request.Headers)
	color.Yellow("Body: %s", request.Body)

	// Save the request data
	if err := s.saveRequest(request); err != nil {
		http.Error(w, "Failed to save request", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Request received and saved")
}

func (s *Server) checkHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read the JSON file to get all saved requests
	data, err := ioutil.ReadFile("requests.json")
	if err != nil {
		http.Error(w, "Failed to read saved requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	server := &Server{}

	// Set up the HTTP handlers
	http.HandleFunc("/", server.requestHandler)
	http.HandleFunc("/check", server.checkHandler)

	// Start the HTTP server
	addr := ":" + *port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
