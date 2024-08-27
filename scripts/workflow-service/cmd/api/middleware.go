package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pascaldekloe/jwt"
	errs "github.com/windevkay/flhoutils/errors"
	"golang.org/x/time/rate"
)

const (
	identityContextKey = contextKey("identityId")
)

type contextKey string

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (app *application) contextSetUser(r *http.Request, identityId int64) *http.Request {
	ctx := context.WithValue(r.Context(), identityContextKey, identityId)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) int64 {
	identityId, ok := r.Context().Value(identityContextKey).(int64)
	if !ok {
		panic("missing identityId value in request context")
	}

	return identityId
}

func (app *application) validateJWTToken(token string) (*jwt.Claims, error) {
	// retrieve public key to decode token
	pubKeyPath := filepath.Join(".", "keys", "ec_public_key.pem")
	pubPem, err := os.ReadFile(pubKeyPath)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(pubPem)
	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	pubKey := pubKeyInterface.(*ecdsa.PublicKey)

	claims, err := jwt.ECDSACheck([]byte(token), pubKey)
	if err != nil {
		return nil, err
	}

	if !claims.Valid(time.Now()) {
		return nil, err
	}

	if claims.Issuer != "github.com/windevkay/flho/identity-service" {
		return nil, err
	}

	if !claims.AcceptAudience("github.com/windevkay/flho") {
		return nil, err
	}

	return claims, nil
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// indicate to caches that the response may vary based on the value of the Authorization
		// header in the request.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			errs.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			errs.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		claims, err := app.validateJWTToken(token)
		if err != nil {
			errs.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		identityId, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			errs.ServerErrorResponse(w, r, err)
			return
		}

		r = app.contextSetUser(r, identityId)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				errs.ServerErrorResponse(w, r, fmt.Errorf("%s", err))
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
				errs.ServerErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				errs.RateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	mw.headerWritten = true
	return mw.wrapped.Write(b)
}

func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

func (app *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_microseconds")
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Add(1)

		mw := newMetricsResponseWriter(w)

		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)

		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
