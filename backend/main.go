package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	sqlite3ConnStr := flag.String("db", "file:db.sqlite3", "Connection string for the SQLite3 database")
	listenAddr := flag.String("listen", ":8080", "Address and port to listen to")
	staticDir := flag.String("static", "static", "Directory to static folder to serve")
	forwardURL := flag.String("forward", "", "URL to forward to, overrides -static flag")
	frontendURL := flag.String("frontend", "http://localhost:8080", "URL of the frontend")
	googleClientID := flag.String("client-id", "", "Google Client ID")
	googleClientSecret := flag.String("client-secret", "", "Google Client Secret")
	stripeSecretKey := flag.String("stripe-secret", "", "Stripe Secret Key")
	stripeWebhookSecret := flag.String("stripe-webhook", "", "Stripe Webhook Secret")
	flag.Parse()

	cfg := &ServerConfig{
		ListenAddr:          *listenAddr,
		Sqlite3ConnStr:      *sqlite3ConnStr,
		FrontendURL:         *frontendURL,
		GoogleClientID:      *googleClientID,
		GoogleClientSecret:  *googleClientSecret,
		StripeSecretKey:     *stripeSecretKey,
		StripeWebhookSecret: *stripeWebhookSecret,
	}
	if *forwardURL != "" {
		parsedURL, err := url.Parse(*forwardURL)
		if err != nil {
			slog.Error("error parsing URL to frorward to", "err", err)
		}
		cfg.Forwarder = httputil.NewSingleHostReverseProxy(parsedURL)
	} else {
		static := http.Dir(*staticDir)
		cfg.StaticDir = &static
	}

	if err := run(cfg); err != nil {
		slog.Error("error running server", "err", err)
		os.Exit(1)
	}
}

type ServerConfig struct {
	ListenAddr     string
	Sqlite3ConnStr string
	StaticDir      *http.Dir
	Forwarder      *httputil.ReverseProxy

	GoogleClientID      string
	GoogleClientSecret  string
	StripeSecretKey     string
	StripeWebhookSecret string
	FrontendURL         string
}

func run(config *ServerConfig) error {
	server, err := NewServer(config)
	if err != nil {
		return fmt.Errorf("error constructing server: %w", err)
	}
	slog.Info("server listening", "addr", config.ListenAddr)
	return http.ListenAndServe(config.ListenAddr, server.HTTPMux())
}
