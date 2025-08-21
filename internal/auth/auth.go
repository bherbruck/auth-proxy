package auth

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// Config holds the authentication configuration
type Config struct {
	Username     string `env:"AUTH_PROXY_USERNAME" flag:"username,u" required:"true" desc:"Username for authentication"`
	Password     string `env:"AUTH_PROXY_PASSWORD" flag:"password,p" desc:"Password for authentication"`
	PasswordHash string `env:"AUTH_PROXY_PASSWORD_HASH" flag:"password-hash" desc:"Bcrypt hash of password (replaces password)"`
	Target       string `env:"AUTH_PROXY_TARGET" flag:"target,t" required:"true" desc:"Target application URL to proxy to"`
	CookieSecret string `env:"AUTH_PROXY_COOKIE_SECRET" flag:"cookie-secret" desc:"Secret key for cookie encryption"`
	LoginTitle   string `env:"AUTH_PROXY_LOGIN_TITLE" flag:"login-title" default:"Auth Proxy" desc:"Custom title for the login page"`
	Port         string `env:"AUTH_PROXY_PORT" flag:"port" default:"8080" desc:"Port to run the auth proxy on"`
}

// AuthProxy handles authentication and proxy functionality
type AuthProxy struct {
	config *Config
	store  *sessions.CookieStore
	proxy  *httputil.ReverseProxy
}

// NewAuthProxy creates a new authentication proxy instance
func NewAuthProxy(config *Config) (*AuthProxy, error) {
	targetURL, err := url.Parse(config.Target)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %v", err)
	}

	store := sessions.NewCookieStore([]byte(config.CookieSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true if using HTTPS
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return &AuthProxy{
		config: config,
		store:  store,
		proxy:  proxy,
	}, nil
}

// GetConfig returns the authentication configuration
func (ap *AuthProxy) GetConfig() *Config {
	return ap.config
}

// GetProxy returns the reverse proxy instance
func (ap *AuthProxy) GetProxy() *httputil.ReverseProxy {
	return ap.proxy
}

// ValidateCredentials validates username and password against configuration
func (ap *AuthProxy) ValidateCredentials(username, password string) bool {
	if username != ap.config.Username {
		return false
	}

	if ap.config.PasswordHash != "" {
		// Use bcrypt hash comparison
		return bcrypt.CompareHashAndPassword([]byte(ap.config.PasswordHash), []byte(password)) == nil
	}

	// Use plain text password comparison
	return password == ap.config.Password
}

// CreateSession creates a new authenticated session for the user
func (ap *AuthProxy) CreateSession(w http.ResponseWriter, r *http.Request, username string) error {
	session, err := ap.store.Get(r, "auth-session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return fmt.Errorf("failed to get session: %v", err)
	}

	session.Values["authenticated"] = true
	session.Values["username"] = username

	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		return fmt.Errorf("failed to save session: %v", err)
	}

	return nil
}

// DestroySession invalidates the current user session
func (ap *AuthProxy) DestroySession(w http.ResponseWriter, r *http.Request) error {
	session, err := ap.store.Get(r, "auth-session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		// Continue with logout even if we can't get the session
	}

	session.Values["authenticated"] = false
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		return fmt.Errorf("failed to destroy session: %v", err)
	}

	return nil
}

// IsAuthenticated checks if the current request has a valid session
func (ap *AuthProxy) IsAuthenticated(r *http.Request) bool {
	session, err := ap.store.Get(r, "auth-session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return false
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	return ok && authenticated
}

// RedirectToLogin redirects unauthenticated requests to the login page
func (ap *AuthProxy) RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	redirectURL := fmt.Sprintf("/auth/login?redirect=%s", url.QueryEscape(r.URL.String()))
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// AuthMiddleware wraps HTTP handlers with authentication checks
func (ap *AuthProxy) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ap.IsAuthenticated(r) {
			ap.RedirectToLogin(w, r)
			return
		}
		next(w, r)
	}
}
