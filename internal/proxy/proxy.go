package proxy

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(10, 20)

func Proxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := godotenv.Load("../../.env")
		if err != nil {
			log.Fatal("Error loading the environment variables")
		}
		secretToken := os.Getenv("SECRET_TOKEN")
		if secretToken == "" {
			log.Fatal("No secret token in the backend")
		}

		token := r.Header.Get("Authorization")
		if token == "" || token != secretToken {
			log.Fatal("Invalid token! Couldn't be authorized")
		}
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			_, err = fmt.Fprintf(w, "Rate Limit Exceeded!! Try in some time")
			if err != nil {
				log.Fatalf("Couldn't write the response: %v", err)
			}
			return
		}
	}
}
