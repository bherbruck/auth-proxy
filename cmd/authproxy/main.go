package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/bherbruck/configlib"
	"github.com/gorilla/mux"

	"github.com/bherbruck/auth-proxy/internal/auth"
	"github.com/bherbruck/auth-proxy/internal/handlers"
)

func main() {
	var config auth.Config

	// Parse configuration using configlib
	if err := configlib.Parse(&config); err != nil {
		log.Fatal("Configuration error:", err)
	}

	// Generate random cookie secret if not provided
	if config.CookieSecret == "" {
		log.Printf("AUTH_PROXY_COOKIE_SECRET not provided, generating random secret")
		config.CookieSecret = generateRandomSecret()
	}

	// Validate that either password or password hash is provided
	if config.Password == "" && config.PasswordHash == "" {
		log.Fatal("Either AUTH_PROXY_PASSWORD or AUTH_PROXY_PASSWORD_HASH is required")
	}

	log.Printf("Starting auth proxy on port %s", config.Port)
	log.Printf("Proxying to: %s", config.Target)

	// Create auth proxy instance
	authProxy, err := auth.NewAuthProxy(&config)
	if err != nil {
		log.Fatal("Failed to create auth proxy:", err)
	}

	// Create HTTP handler
	handler := handlers.NewHandler(authProxy)

	// Setup routes
	router := mux.NewRouter()

	// Auth routes
	router.HandleFunc("/auth/login", handler.HandleLogin).Methods("GET", "POST")
	router.HandleFunc("/auth/logout", handler.HandleLogout).Methods("GET", "POST")

	// Static files for React app (all files including assets/ subdirectory)
	router.PathPrefix("/static/").HandlerFunc(handler.HandleStaticFiles)

	// All other routes go through the auth middleware and then proxy
	router.PathPrefix("/").HandlerFunc(authProxy.AuthMiddleware(handler.HandleProxy))

	log.Fatal(http.ListenAndServe(":"+config.Port, router))
}

func generateRandomSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Failed to generate random secret:", err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}
