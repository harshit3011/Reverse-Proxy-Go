package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	loadbalancing "github.com/harshit3011/Reverse-Proxy-Go/internal/load_balancing"
	"github.com/harshit3011/Reverse-Proxy-Go/internal/proxy"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Could not load .env file: %v", err)
	}
	if os.Getenv("SECRET_TOKEN") == "" {
		log.Fatal("SECRET_TOKEN is not set")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	addresses := strings.Split(os.Getenv("ORIGIN_SERVERS"), ",")
	if len(addresses) == 1 && addresses[0] == "" {
		addresses = make([]string, 5)
		for i := range addresses {
			addresses[i] = fmt.Sprintf("http://127.0.0.1:800%d", i)
		}
	}
	originServers := make([]*url.URL, len(addresses))
	for i, address := range addresses {
		parsedURL, err := url.Parse(strings.TrimSpace(address))
		if err != nil {
			log.Fatalf("Couldn't parse the given address: %v", err)
		}
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			log.Fatalf("Origin server address must include a scheme and host: %q", address)
		}
		originServers[i] = parsedURL
	}
	lb := loadbalancing.NewLoadBalancer(originServers)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: proxy.Proxy(lb),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server couldn't be connected!!")
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server...")
}
