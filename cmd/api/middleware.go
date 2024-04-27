package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func cleanupClientRateMap(clients *map[string]*client, mu *sync.Mutex) {
	for {
		time.Sleep(time.Minute)

		mu.Lock()

		for ip, client := range *clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(*clients, ip)
			}
		}

		mu.Unlock()
	}
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// everything out here is reused per request, not re created
	// utilize mutexes as maps arent concurrency safe
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// cleanup background job for clients map
	go cleanupClientRateMap(&clients, &mu)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}
