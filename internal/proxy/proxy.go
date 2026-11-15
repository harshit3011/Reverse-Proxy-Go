package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	loadbalancing "github.com/harshit3011/Reverse-Proxy-Go/internal/load_balancing"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(10, 20)

func Proxy(lb *loadbalancing.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secretToken := os.Getenv("SECRET_TOKEN")
		if secretToken == "" {
			http.Error(w, "Proxy authentication is not configured", http.StatusInternalServerError)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" || token != secretToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded; try again later", http.StatusTooManyRequests)
			return
		}

		if lb == nil {
			http.Error(w, "No origin server is available", http.StatusServiceUnavailable)
			return
		}

		serverURL, err := lb.GetNextServer()
		if err != nil {
			log.Printf("Could not select an origin server: %v", err)
			http.Error(w, "No origin server is available", http.StatusServiceUnavailable)
			return
		}
		log.Printf("Forwarding request to %s at %s", serverURL, time.Now())

		targetURL := *serverURL
		targetURL.Path = r.URL.Path
		targetURL.RawPath = r.URL.RawPath
		targetURL.RawQuery = r.URL.RawQuery

		outgoingRequest, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL.String(), r.Body)
		if err != nil {
			http.Error(w, "Could not create upstream request", http.StatusInternalServerError)
			return
		}
		outgoingRequest.ContentLength = r.ContentLength
		outgoingRequest.Header = r.Header.Clone()
		removeHopByHopHeaders(outgoingRequest.Header)

		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if prior := outgoingRequest.Header.Get("X-Forwarded-For"); prior != "" {
				clientIP = prior + ", " + clientIP
			}
			outgoingRequest.Header.Set("X-Forwarded-For", clientIP)
		}
		originServerResponse, err := http.DefaultClient.Do(outgoingRequest)
		if err != nil {
			log.Printf("Could not forward request to %s: %v", serverURL, err)
			http.Error(w, "Bad gateway", http.StatusBadGateway)
			return
		}
		defer originServerResponse.Body.Close()

		for header, values := range originServerResponse.Header {
			for _, value := range values {
				w.Header().Add(header, value)
			}
		}
		removeHopByHopHeaders(w.Header())
		w.WriteHeader(originServerResponse.StatusCode)
		if _, err := io.Copy(w, originServerResponse.Body); err != nil {
			log.Printf("Could not copy origin response body: %v", err)
		}
	}
}

func removeHopByHopHeaders(header http.Header) {
	for _, connectionHeader := range header.Values("Connection") {
		for _, name := range strings.Split(connectionHeader, ",") {
			header.Del(strings.TrimSpace(name))
		}
	}
	for _, name := range []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	} {
		header.Del(name)
	}
}
