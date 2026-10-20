package web

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/luis-octavius/akrasia/internal/db"
	database "github.com/luis-octavius/akrasia/internal/db/out"
	"github.com/luis-octavius/akrasia/internal/tasks"
)

const testToken = "0123456789abcdef0123456789abcdef"

// newTestServer serves a migrated temp database (DB_PATH points at it).
func newTestServer(t *testing.T) (*Server, *tasks.TaskManager) {
	t.Helper()
	t.Setenv("DB_PATH", filepath.Join(t.TempDir(), "akrasia.db"))

	conn, err := db.InitDB()
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	tkm := &tasks.TaskManager{Queries: database.New(conn)}
	s, err := New(tkm, testToken)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return s, tkm
}

// do sends a request; logged in adds the token cookie, htmx the header htmx
// sends with every request.
func do(t *testing.T, s *Server, method, target string, form url.Values, loggedIn, htmx bool) *httptest.ResponseRecorder {
	t.Helper()

	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	r := httptest.NewRequest(method, target, body)
	if form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if loggedIn {
		r.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	}
	if htmx {
		r.Header.Set("HX-Request", "true")
	}

	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func addTask(t *testing.T, tkm *tasks.TaskManager, in tasks.NewTodo) {
	t.Helper()
	if in.ExpiresAt.IsZero() && in.Parent == "" {
		in.ExpiresAt = time.Now().AddDate(0, 0, 2)
	}
	if _, err := tkm.AddTodo(context.Background(), in); err != nil {
		t.Fatalf("AddTodo(%q) error = %v", in.Name, err)
	}
}

func TestPagesNeedTheToken(t *testing.T) {
	s, _ := newTestServer(t)

	w := do(t, s, "GET", "/today", nil, false, false)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Errorf("GET /today without token = %d %q, want a redirect to /login", w.Code, w.Header().Get("Location"))
	}

	w = do(t, s, "POST", "/add", url.Values{"name": {"x"}}, false, true)
	if w.Code != http.StatusUnauthorized || w.Header().Get("HX-Redirect") != "/login" {
		t.Errorf("htmx POST without token = %d, HX-Redirect %q", w.Code, w.Header().Get("HX-Redirect"))
	}

	// scripts can send the token as a bearer header
	r := httptest.NewRequest("GET", "/today", nil)
	r.Header.Set("Authorization", "Bearer "+testToken)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("GET /today with a bearer token = %d", w.Code)
	}

	// what installs the app needs no login, and holds no data
	for _, path := range []string{"/manifest.json", "/sw.js", "/app.css", "/app.js", "/htmx.min.js", "/icon-192.png", "/login"} {
		if w := do(t, s, "GET", path, nil, false, false); w.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200 without login", path, w.Code)
		}
	}
}

func TestLogin(t *testing.T) {
	s, _ := newTestServer(t)

	w := do(t, s, "POST", "/login", url.Values{"token": {"wrong"}}, false, false)
	if w.Code != http.StatusUnauthorized || len(w.Result().Cookies()) != 0 {
		t.Fatalf("wrong token = %d with cookies %v", w.Code, w.Result().Cookies())
	}

	w = do(t, s, "POST", "/login", url.Values{"token": {" " + testToken + " "}}, false, false)
	cookies := w.Result().Cookies()
	if w.Code != http.StatusSeeOther || len(cookies) != 1 {
		t.Fatalf("right token = %d with cookies %v", w.Code, cookies)
	}
	c := cookies[0]
	if c.Value != testToken || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie = %+v; want the token, HttpOnly and SameSite=Strict", c)
	}
}

func TestPagesListTasks(t *testing.T) {
	s, tkm := newTestServer(t)
	addTask(t, tkm, tasks.NewTodo{Name: "Party"})
	addTask(t, tkm, tasks.NewTodo{Name: "Bake cake", Parent: "Party"})
	addTask(t, tkm, tasks.NewTodo{Name: "Run", IsDaily: true, ExpiresAt: time.Now().AddDate(0, 0, 1)})

	w := do(t, s, "GET", "/all", nil, true, false)
	body := w.Body.String()
	for _, want := range []string{"<!doctype html>", "Party", "Bake cake", "0/1", `class="todo sub"`, "/manifest.json"} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /all misses %q", want)
		}
	}

	// the Today tab lists the pending daily task; htmx gets only the app
	w = do(t, s, "GET", "/today", nil, true, true)
	body = w.Body.String()
	if strings.Contains(body, "<!doctype html>") || !strings.Contains(body, `<main id="app">`) {
		t.Errorf("htmx GET /today should be the app alone:\n%.200s", body)
	}
	if !strings.Contains(body, "Run") {
		t.Errorf("GET /today misses the daily task")
	}

	if csp := w.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("Content-Security-Policy = %q", csp)
	}
}

func TestNamesAreEscaped(t *testing.T) {
	s, tkm := newTestServer(t)
	addTask(t, tkm, tasks.NewTodo{Name: `<script>alert("x")</script>`})

	body := do(t, s, "GET", "/all", nil, true, false).Body.String()
	if strings.Contains(body, "<script>alert") {
		t.Fatal("a task name was written into the page unescaped")
	}
}

func TestWritesNeedHtmx(t *testing.T) {
	s, tkm := newTestServer(t)

	// a form posted from another site can't add the HX-Request header
	w := do(t, s, "POST", "/add", url.Values{"name": {"Sneaky"}}, true, false)
	if w.Code != http.StatusForbidden {
		t.Errorf("POST /add without HX-Request = %d, want 403", w.Code)
	}
	if todos, _ := tkm.ListTodos(context.Background(), ""); len(todos) != 0 {
		t.Errorf("a task was added: %v", todos)
	}
}

func TestAdd(t *testing.T) {
	s, tkm := newTestServer(t)
	ctx := context.Background()
	addTask(t, tkm, tasks.NewTodo{Name: "Party", ExpiresAt: time.Now().AddDate(0, 0, 5)})

	w := do(t, s, "POST", "/add", url.Values{"name": {"Pay rent"}, "kind": {"task"}, "tab": {"all"}}, true, true)
	if !strings.Contains(w.Body.String(), "Pay rent") || !strings.Contains(w.Body.String(), `class="msg ok"`) {
		t.Errorf("POST /add should show the new task and a status:\n%s", w.Body.String())
	}
	rent, err := tkm.FindTodo(ctx, "Pay rent")
	if err != nil || rent.ExpiresAt.Format(time.DateOnly) != time.Now().AddDate(0, 0, 1).Format(time.DateOnly) {
		t.Errorf("Pay rent = %v, %v; want it due tomorrow", rent.ExpiresAt, err)
	}

	do(t, s, "POST", "/add", url.Values{"name": {"Bake cake"}, "kind": {"task"}, "parent": {"Party"}}, true, true)
	if cake, err := tkm.FindTodo(ctx, "Party/Bake cake"); err != nil || cake.Parent != "Party" {
		t.Errorf("subtask = %+v, %v", cake, err)
	}

	date := time.Now().AddDate(0, 0, 3).Format(time.DateOnly)
	do(t, s, "POST", "/add", url.Values{"name": {"Dentist"}, "kind": {"task"}, "date": {date}, "priority": {"high"}}, true, true)
	if dentist, err := tkm.FindTodo(ctx, "Dentist"); err != nil || dentist.ExpiresAt.Format(time.DateOnly) != date || dentist.Priority != "high" {
		t.Errorf("Dentist = %+v, %v; want due %s, high", dentist, err, date)
	}

	do(t, s, "POST", "/add", url.Values{"name": {"Idea"}, "kind": {"dump"}}, true, true)
	if dumps, _ := tkm.Dumps(ctx); len(dumps) != 1 || dumps[0].Name != "Idea" {
		t.Errorf("dumps = %v", dumps)
	}

	// the service's errors come back on the page
	past := time.Now().AddDate(0, 0, -3).Format(time.DateOnly)
	w = do(t, s, "POST", "/add", url.Values{"name": {"Late"}, "kind": {"task"}, "date": {past}}, true, true)
	if !strings.Contains(w.Body.String(), `class="msg err"`) {
		t.Errorf("a past date should show an error:\n%s", w.Body.String())
	}
}

func TestCompleteParentAsksAndCompletesAll(t *testing.T) {
	s, tkm := newTestServer(t)
	ctx := context.Background()
	addTask(t, tkm, tasks.NewTodo{Name: "Party"})
	addTask(t, tkm, tasks.NewTodo{Name: "Bake cake", Parent: "Party"})
	addTask(t, tkm, tasks.NewTodo{Name: "Invite friends", Parent: "Party"})

	w := do(t, s, "GET", "/complete?path=Party&tab=all", nil, true, true)
	if !strings.Contains(w.Body.String(), `name="all"`) || !strings.Contains(w.Body.String(), "2") {
		t.Fatalf("dialog for a parent should ask about its 2 pending subtasks:\n%s", w.Body.String())
	}

	do(t, s, "POST", "/complete", url.Values{"path": {"Party"}, "all": {"1"}, "notes": {"great"}, "tab": {"all"}}, true, true)
	party, err := tkm.FindTodo(ctx, "Party")
	if err != nil || !party.Done() || party.SubtaskProgress().Done != 2 {
		t.Errorf("after completing all: %+v, %v", party, err)
	}
}

func TestCompleteSubtask(t *testing.T) {
	s, tkm := newTestServer(t)
	addTask(t, tkm, tasks.NewTodo{Name: "Party"})
	addTask(t, tkm, tasks.NewTodo{Name: "Bake cake", Parent: "Party"})

	w := do(t, s, "GET", "/complete?path=Party%2FBake+cake&tab=today", nil, true, true)
	if strings.Contains(w.Body.String(), `name="all"`) || !strings.Contains(w.Body.String(), "Bake cake") {
		t.Fatalf("dialog for a subtask:\n%s", w.Body.String())
	}

	do(t, s, "POST", "/complete", url.Values{"path": {"Party/Bake cake"}, "tab": {"today"}}, true, true)
	if party, _ := tkm.FindTodo(context.Background(), "Party"); !party.Done() {
		t.Error("the parent should be done with its only subtask")
	}

	if w := do(t, s, "GET", "/complete?path=Nope", nil, true, true); w.Code != http.StatusNotFound {
		t.Errorf("dialog for a missing task = %d, want 404", w.Code)
	}
}

func TestIcons(t *testing.T) {
	s, _ := newTestServer(t)

	for _, size := range []int{192, 512} {
		w := do(t, s, "GET", "/icon-"+strconv.Itoa(size)+".png", nil, false, false)
		img, err := png.Decode(bytes.NewReader(w.Body.Bytes()))
		if err != nil || img.Bounds().Dx() != size {
			t.Errorf("icon %d: %v, %v", size, img, err)
		}
	}
	if w := do(t, s, "GET", "/icon-64.png", nil, false, false); w.Code != http.StatusNotFound {
		t.Errorf("unknown icon size = %d", w.Code)
	}
}

func TestLock(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "akrasia.db")

	if Running(dbPath) {
		t.Fatal("Running() before any Lock")
	}
	release, err := Lock(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if !Running(dbPath) {
		t.Error("Running() = false while locked")
	}
	if _, err := Lock(dbPath); !errors.Is(err, ErrRunning) {
		t.Errorf("second Lock() error = %v, want ErrRunning", err)
	}

	release()
	if Running(dbPath) {
		t.Error("Running() = true after release")
	}
}

func TestConfigCreatesAPrivateToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := LoadConfig()
	if err != nil || len(cfg.Token) != 32 || cfg.ListenAddr() != DefaultAddr {
		t.Fatalf("LoadConfig() = %+v, %v", cfg, err)
	}
	again, _ := LoadConfig()
	if again.Token != cfg.Token {
		t.Error("the token changed between loads")
	}

	info, err := os.Stat(configFile())
	if err != nil || info.Mode().Perm() != 0600 {
		t.Errorf("serve.json mode = %v, %v; want 0600", info.Mode(), err)
	}
}
