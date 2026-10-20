package web

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

type tab string

const (
	tabToday tab = "today"
	tabAll   tab = "all"
	tabDump  tab = "dump"
)

func parseTab(s string) tab {
	switch tab(s) {
	case tabAll, tabDump:
		return tab(s)
	}
	return tabToday
}

var funcs = template.FuncMap{
	"T":   i18n.T,
	"due": func(t time.Time) string { return t.Format("02/01") },
	// stamp is when a dump entry was written (stored in local time)
	"stamp": func(t time.Time) string { return t.Format("02/01 15:04") },
}

type loginData struct {
	Lang   string
	Failed bool
}

// pageData is everything a tab shows.
type pageData struct {
	Lang     string
	Tab      tab
	Today    int // tasks on the Today tab
	All      int // top-level tasks
	Dump     int // dump entries
	Progress tasks.Progress
	Percent  int
	Sections []section
	// Dumps fill the Dump tab
	Dumps []tasks.DumpEntry
	// Parents are the top-level tasks a new subtask can go under
	Parents []string
	Status  string
	Error   string
	// DateMin is today, the earliest due date the add form accepts
	DateMin string
}

type section struct {
	Title string
	Rows  []row
}

// row is a task as a template sees it, with the tab its buttons come back
// to.
type row struct {
	tasks.TodoView
	Tab tab
}

func rows(t tab, todos []tasks.TodoView) []row {
	out := make([]row, len(todos))
	for i, todo := range todos {
		out[i] = row{todo, t}
	}
	return out
}

// Subs are the subtask rows under a parent.
func (r row) Subs() []row { return rows(r.Tab, r.Subtasks) }

// Late reports whether a one-off task is past its due day.
func (r row) Late() bool {
	return !r.IsDaily && !r.Done() && r.ExpiresAt.Before(dayStart(time.Now()))
}

func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// completeData is the dialog that completes a task.
type completeData struct {
	Tab     tab
	Todo    tasks.TodoView
	Pending int // pending subtasks the dialog would complete too
}

func (s *Server) load(r *http.Request, t tab) (pageData, error) {
	ctx := r.Context()
	data := pageData{Lang: i18n.GetCurrentLanguage(), Tab: t, DateMin: s.now().Format(time.DateOnly)}

	view, err := s.svc.Today(ctx, tasks.TodayOptions{})
	if err != nil {
		return data, err
	}
	all, err := s.svc.ListTodos(ctx, "")
	if err != nil {
		return data, err
	}
	if data.Progress, err = s.svc.TodayProgress(ctx); err != nil {
		return data, err
	}
	dumps, err := s.svc.Dumps(ctx)
	if err != nil {
		return data, err
	}
	data.Dump = len(dumps)
	data.Percent = int(data.Progress.Ratio() * 100)

	todaySections := []section{
		{i18n.T("tasks.tasks.overdue"), rows(t, view.Overdue)},
		{i18n.T("tasks.tasks.dueToday"), rows(t, view.DueToday)},
		{i18n.T("tasks.tasks.dailyPending"), rows(t, view.Daily)},
		{i18n.T("tasks.tasks.expiringSoon"), rows(t, view.ExpiringSoon)},
	}
	for _, sec := range todaySections {
		data.Today += len(sec.Rows)
	}
	data.All = len(all)

	switch t {
	case tabDump:
		data.Dumps = dumps
	case tabAll:
		if len(all) > 0 {
			data.Sections = []section{{Rows: rows(t, all)}}
		}
	default:
		for _, sec := range todaySections {
			if len(sec.Rows) > 0 {
				data.Sections = append(data.Sections, sec)
			}
		}
	}

	for _, todo := range all {
		data.Parents = append(data.Parents, todo.Name)
	}
	return data, nil
}

// page serves a tab: the whole page, or only the app when htmx switches
// tabs (htmx asks for the whole page when it restores history).
func (s *Server) page(t tab) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := s.load(r, t)
		if err != nil {
			data.Error = err.Error()
		}

		name := "page"
		if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-History-Restore-Request") != "true" {
			name = "app"
		}
		s.render(w, name, data)
	}
}

// reply re-renders the app of tab after a write, with its outcome.
func (s *Server) reply(w http.ResponseWriter, r *http.Request, t tab, status string, failure error) {
	data, err := s.load(r, t)
	if err != nil {
		failure = errors.Join(failure, err)
	}
	if failure != nil {
		data.Error = failure.Error()
	} else {
		data.Status = status
	}
	s.render(w, "app", data)
}

// add creates a task, a subtask (with a parent) or a dump entry.
func (s *Server) add(w http.ResponseWriter, r *http.Request) {
	t := parseTab(r.FormValue("tab"))
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		s.reply(w, r, t, "", errors.New(i18n.T("tui.form.nameRequired")))
		return
	}

	if r.FormValue("kind") == "dump" {
		entry, err := s.svc.AddDump(r.Context(), name)
		s.reply(w, r, t, fmt.Sprintf(i18n.T("tui.status.dumpAdded"), entry.Name), err)
		return
	}

	in := tasks.NewTodo{
		Name:     name,
		Parent:   strings.TrimSpace(r.FormValue("parent")),
		Priority: r.FormValue("priority"),
		IsDaily:  r.FormValue("daily") != "",
	}

	due, err := parseDate(r.FormValue("date"), in.Parent != "", s.now())
	if err != nil {
		s.reply(w, r, t, "", err)
		return
	}
	in.ExpiresAt = due

	todo, err := s.svc.AddTodo(r.Context(), in)
	s.reply(w, r, t, fmt.Sprintf(i18n.T("tui.status.created"), todo.Path()), err)
}

// parseDate reads the form's yyyy-mm-dd date with the rules of
// `akrasia add --date`. No date means tomorrow, or the parent's date for a
// subtask (the zero time).
func parseDate(value string, subtask bool, now time.Time) (time.Time, error) {
	if value == "" {
		if subtask {
			return time.Time{}, nil
		}
		return tasks.ParseDateAt(nil, now)
	}

	d, err := time.ParseInLocation(time.DateOnly, value, now.Location())
	if err != nil {
		return time.Time{}, errors.New(i18n.T("tui.form.invalidDate"))
	}
	return tasks.ParseDateAt([]int{d.Day(), int(d.Month()), d.Year()}, now)
}

// completeForm is the dialog that asks for the notes, and for a parent
// with pending subtasks, whether to complete them too.
func (s *Server) completeForm(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	t := parseTab(r.FormValue("tab"))

	all, err := s.svc.ListTodos(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := completeData{Tab: t}
	found := false
	for _, todo := range all {
		if todo.Path() == path {
			data.Todo, found = todo, true
			p := todo.SubtaskProgress()
			data.Pending = p.Total - p.Done
		}
		for _, sub := range todo.Subtasks {
			if sub.Path() == path {
				data.Todo, found = sub, true
			}
		}
	}
	if !found {
		http.Error(w, fmt.Sprintf(i18n.T("tasks.error.todoNotFound"), path), http.StatusNotFound)
		return
	}

	s.render(w, "complete", data)
}

func (s *Server) complete(w http.ResponseWriter, r *http.Request) {
	t := parseTab(r.FormValue("tab"))
	path, notes := r.FormValue("path"), strings.TrimSpace(r.FormValue("notes"))

	var (
		todo tasks.TodoView
		err  error
	)
	if r.FormValue("all") != "" {
		todo, err = s.svc.CompleteAll(r.Context(), path, notes)
	} else {
		todo, err = s.svc.Complete(r.Context(), path, notes)
	}
	s.reply(w, r, t, fmt.Sprintf(i18n.T("tui.status.completed"), todo.Name), err)
}

// convertData is the dialog that turns a dump entry into a task.
type convertData struct {
	Dump    tasks.DumpEntry
	Parents []string
	DateMin string
}

// findDump looks up the entry a form names by id, so a stale page can't
// act on an entry that is gone.
func (s *Server) findDump(r *http.Request) (tasks.DumpEntry, error) {
	gone := errors.New(i18n.T("web.dumpGone"))
	id, err := uuid.Parse(r.FormValue("id"))
	if err != nil {
		return tasks.DumpEntry{}, gone
	}
	dumps, err := s.svc.Dumps(r.Context())
	if err != nil {
		return tasks.DumpEntry{}, err
	}
	for _, d := range dumps {
		if d.ID == id {
			return d, nil
		}
	}
	return tasks.DumpEntry{}, gone
}

func (s *Server) deleteDump(w http.ResponseWriter, r *http.Request) {
	entry, err := s.findDump(r)
	if err == nil {
		err = s.svc.DeleteDump(r.Context(), entry.ID)
	}
	s.reply(w, r, tabDump, fmt.Sprintf(i18n.T("tui.status.dumpDeleted"), entry.Name), err)
}

// convertForm is the dialog that creates a task from a dump entry, with
// its text as the name.
func (s *Server) convertForm(w http.ResponseWriter, r *http.Request) {
	entry, err := s.findDump(r)
	if err != nil {
		http.Error(w, i18n.T("web.dumpGone"), http.StatusNotFound)
		return
	}

	all, err := s.svc.ListTodos(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := convertData{Dump: entry, DateMin: s.now().Format(time.DateOnly)}
	for _, todo := range all {
		data.Parents = append(data.Parents, todo.Name)
	}
	s.render(w, "convert", data)
}

// convert creates the task and removes the entry from the dump; when the
// task can't be created the entry stays.
func (s *Server) convert(w http.ResponseWriter, r *http.Request) {
	entry, err := s.findDump(r)
	if err != nil {
		s.reply(w, r, tabDump, "", err)
		return
	}

	in := tasks.NewTodo{
		Name:     strings.TrimSpace(r.FormValue("name")),
		Parent:   strings.TrimSpace(r.FormValue("parent")),
		Priority: r.FormValue("priority"),
		IsDaily:  r.FormValue("daily") != "",
	}
	if in.Name == "" {
		s.reply(w, r, tabDump, "", errors.New(i18n.T("tui.form.nameRequired")))
		return
	}
	if in.ExpiresAt, err = parseDate(r.FormValue("date"), in.Parent != "", s.now()); err != nil {
		s.reply(w, r, tabDump, "", err)
		return
	}

	todo, err := s.svc.PromoteDump(r.Context(), entry.ID, in)
	s.reply(w, r, tabDump, fmt.Sprintf(i18n.T("tui.status.promoted"), todo.Path()), err)
}
