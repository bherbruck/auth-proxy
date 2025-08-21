package handlers

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bherbruck/auth-proxy/internal/auth"
	"github.com/bherbruck/auth-proxy/ui"
)

// Handler manages HTTP requests for the auth proxy
type Handler struct {
	authProxy *auth.AuthProxy
}

// NewHandler creates a new HTTP handler instance
func NewHandler(authProxy *auth.AuthProxy) *Handler {
	return &Handler{
		authProxy: authProxy,
	}
}

// HandleLogin processes both GET and POST requests for the login page
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.renderLogin(w, "", nil)
		return
	}

	// POST method - process login
	username := r.FormValue("username")
	password := r.FormValue("password")

	if !h.authProxy.ValidateCredentials(username, password) {
		h.renderLogin(w, "Invalid username or password", nil)
		return
	}

	// Create session
	if err := h.authProxy.CreateSession(w, r, username); err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to original URL or root
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleLogout processes logout requests and destroys the session
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.authProxy.DestroySession(w, r); err != nil {
		log.Printf("Error destroying session: %v", err)
		// Continue with logout even if session destruction fails
	}

	http.Redirect(w, r, "/auth/login", http.StatusFound)
}

// HandleProxy forwards authenticated requests to the target application
func (h *Handler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	h.authProxy.GetProxy().ServeHTTP(w, r)
}

// HandleStaticFiles serves the built React app static files
func (h *Handler) HandleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Remove /static prefix from path
	path := strings.TrimPrefix(r.URL.Path, "/static")
	if path == "" || path == "/" {
		path = "/index.html"
	}

	file, err := ui.DistDirFS.Open(strings.TrimPrefix(path, "/"))
	if err != nil {
		log.Printf("Error opening static file %s: %v", path, err)
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading static file %s: %v", path, err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Determine content type
	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", http.DetectContentType(content))
	}

	w.Write(content)
}

// renderLogin serves the React login app with server-side injected variables
func (h *Handler) renderLogin(w http.ResponseWriter, errorMsg string, data map[string]interface{}) {
	// Read the built index.html from embedded FS
	file, err := ui.DistDirFS.Open("index.html")
	if err != nil {
		log.Printf("Error opening index.html: %v", err)
		http.Error(w, "Login page not found", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading index.html: %v", err)
		http.Error(w, "Error loading login page", http.StatusInternalServerError)
		return
	}

	// Inject config variables as JavaScript in the head
	configScript := `<script>
		window.AUTH_CONFIG = {
			title: "` + template.JSEscapeString(h.authProxy.GetConfig().LoginTitle) + `",
			error: "` + template.JSEscapeString(errorMsg) + `"
		};
	</script>`

	// Insert before closing </head> tag
	htmlContent := string(content)
	htmlContent = strings.Replace(htmlContent, "</head>", configScript+"</head>", 1)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent))
}
