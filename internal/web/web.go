// Package web is the browser interface of `akrasia serve`: pages rendered
// by Go templates, with htmx for the clicks, made to be installed on a
// phone's home screen. It reads and writes through the tasks service, the
// same one the CLI and the TUI use.
package web

import (
	"context"
	"crypto/subtle"
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// ErrRunning is returned by Lock when `serve` already uses the database.
var ErrRunning = errors.New("akrasia serve is already running on this database")

// Service is the part of tasks.TaskManager the web interface uses.
type Service interface {
	Today(ctx context.Context, opts tasks.TodayOptions) (tasks.TodayView, error)
	ListTodos(ctx context.Context, priorityFilter string) ([]tasks.TodoView, error)
	TodayProgress(ctx context.Context) (tasks.Progress, error)
	AddTodo(ctx context.Context, in tasks.NewTodo) (tasks.TodoView, error)
	AddDump(ctx context.Context, name string) (tasks.DumpEntry, error)
	Complete(ctx context.Context, name, notes string) (tasks.TodoView, error)
	CompleteAll(ctx context.Context, name, notes string) (tasks.TodoView, error)
}

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

// cookieName holds the token once a device has logged in.
const cookieName = "akrasia_token"

// Server is the web interface's http.Handler.
type Server struct {
	svc   Service
	token string
	pages *template.Template
	mux   *http.ServeMux
	// now is time.Now; tests fix it
	now func() time.Time
}

// New builds the handler. Every page needs token, given once on the login
// page and kept in a cookie.
func New(svc Service, token string) (*Server, error) {
	pages, err := template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	s := &Server{svc: svc, token: token, pages: pages, mux: http.NewServeMux(), now: time.Now}

	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	files := http.FileServerFS(static)
	for _, name := range []string{"/htmx.min.js", "/app.js", "/app.css", "/manifest.json", "/sw.js"} {
		s.mux.Handle("GET "+name, files)
	}
	s.mux.HandleFunc("GET /icon-192.png", s.icon)
	s.mux.HandleFunc("GET /icon-512.png", s.icon)
	s.mux.HandleFunc("GET /login", s.loginPage)
	s.mux.HandleFunc("POST /login", s.login)

	s.mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/today", http.StatusSeeOther)
	})
	s.mux.Handle("GET /today", s.private(s.page(tabToday)))
	s.mux.Handle("GET /all", s.private(s.page(tabAll)))
	s.mux.Handle("POST /add", s.private(s.mutation(s.add)))
	s.mux.Handle("GET /complete", s.private(http.HandlerFunc(s.completeForm)))
	s.mux.Handle("POST /complete", s.private(s.mutation(s.complete)))
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; frame-ancestors 'none'")
	s.mux.ServeHTTP(w, r)
}

// private lets a request through with the token in its cookie (or an
// Authorization: Bearer header, for scripts); anything else goes to the
// login page.
func (s *Server) private(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.authorized(r) {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}

func (s *Server) authorized(r *http.Request) bool {
	var given string
	if c, err := r.Cookie(cookieName); err == nil {
		given = c.Value
	}
	if auth, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		given = auth
	}
	return given != "" && subtle.ConstantTimeCompare([]byte(given), []byte(s.token)) == 1
}

// lockFile sits next to the database it guards.
func lockFile(dbPath string) string { return dbPath + ".serve.lock" }

// mutation only accepts requests sent by htmx: another site can make a
// browser post a form, but not add a custom header without CORS, which is
// never allowed here. With the SameSite cookie, that closes CSRF.
func (s *Server) mutation(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") != "true" {
			http.Error(w, "htmx requests only", http.StatusForbidden)
			return
		}
		next(w, r)
	})
}

func (s *Server) loginPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "login", loginData{Lang: i18n.GetCurrentLanguage()})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.FormValue("token"))
	if subtle.ConstantTimeCompare([]byte(token), []byte(s.token)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		s.render(w, "login", loginData{Lang: i18n.GetCurrentLanguage(), Failed: true})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    s.token,
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		// behind `tailscale serve` the browser speaks HTTPS to Tailscale
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/today", http.StatusSeeOther)
}

// render writes a template; a failure halfway can only be logged.
func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.pages.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("web: rendering %s: %v", name, err)
	}
}
