package notify

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

type fakeService struct {
	reminders tasks.Reminders
	withDaily []bool
	sent      map[string][]string
}

func (f *fakeService) Reminders(_ context.Context, _ time.Time, _ int, withDaily bool) (tasks.Reminders, error) {
	f.withDaily = append(f.withDaily, withDaily)
	r := f.reminders
	if !withDaily {
		r.Daily = nil
	}
	return r, nil
}

func (f *fakeService) SentSlots(_ context.Context, day string) ([]string, error) {
	return f.sent[day], nil
}

func (f *fakeService) MarkSlotsSent(_ context.Context, day string, slots []string) error {
	if f.sent == nil {
		f.sent = map[string][]string{}
	}
	f.sent[day] = append(f.sent[day], slots...)
	return nil
}

// recorder is a Sender that keeps what it was asked to send.
type recorder struct {
	messages []string
	err      error
}

func (r *recorder) send(_ context.Context, content string) error {
	if r.err != nil {
		return r.err
	}
	r.messages = append(r.messages, content)
	return nil
}

var testCfg = Config{Webhook: "https://discord.com/api/webhooks/1/x", Mention: "123456789012345678"}

func at(hour, minute int) time.Time {
	return time.Date(2026, 10, 19, hour, minute, 0, 0, time.Local)
}

func pending() tasks.Reminders {
	return tasks.Reminders{
		DueSoon: []tasks.TodoView{{Name: "Report", ExpiresAt: at(23, 59)}},
		Daily:   []tasks.TodoView{{Name: "Run", IsDaily: true}},
	}
}

func TestRunSendsEachTimeOnce(t *testing.T) {
	svc := &fakeService{reminders: pending()}
	rec := &recorder{}
	ctx := context.Background()

	for _, step := range []struct {
		now       time.Time
		wantSent  bool
		withDaily bool
	}{
		{at(9, 59), false, false},
		{at(10, 0), true, false},
		{at(10, 15), false, false},
		{at(19, 5), true, true},
		{at(23, 0), false, false},
	} {
		svc.withDaily = nil
		sent, err := Run(ctx, svc, testCfg, step.now, rec.send)
		if err != nil {
			t.Fatalf("%s: Run() error = %v", step.now.Format("15:04"), err)
		}
		if sent != step.wantSent {
			t.Errorf("%s: sent = %v, want %v", step.now.Format("15:04"), sent, step.wantSent)
		}
		if step.wantSent && !slices.Equal(svc.withDaily, []bool{step.withDaily}) {
			t.Errorf("%s: daily tasks asked = %v, want %v", step.now.Format("15:04"), svc.withDaily, step.withDaily)
		}
	}

	if len(rec.messages) != 2 {
		t.Fatalf("sent %d messages in a day, want 2", len(rec.messages))
	}
	if strings.Contains(rec.messages[0], "Run") || !strings.Contains(rec.messages[1], "Run") {
		t.Errorf("daily tasks must be only in the 19:00 message:\n%s\n---\n%s", rec.messages[0], rec.messages[1])
	}
}

func TestRunCatchesUpMissedTimesInOneMessage(t *testing.T) {
	svc := &fakeService{reminders: pending()}
	rec := &recorder{}

	// the machine was off at 10:00 and 19:00
	if _, err := Run(context.Background(), svc, testCfg, at(21, 0), rec.send); err != nil {
		t.Fatal(err)
	}

	if len(rec.messages) != 1 || !strings.Contains(rec.messages[0], "Run") {
		t.Errorf("messages = %q, want one with the daily tasks", rec.messages)
	}
	if got := svc.sent["2026-10-19"]; !slices.Equal(got, []string{"10:00", "19:00"}) {
		t.Errorf("marked %v, want both times", got)
	}
}

func TestRunWithNothingToRemindStaysQuiet(t *testing.T) {
	svc := &fakeService{}
	rec := &recorder{}

	sent, err := Run(context.Background(), svc, testCfg, at(10, 0), rec.send)
	if err != nil || sent || len(rec.messages) != 0 {
		t.Fatalf("Run() = %v, %v with %d messages; want nothing sent", sent, err, len(rec.messages))
	}
	if len(svc.sent["2026-10-19"]) != 1 {
		t.Error("an empty time must still be marked, or it would be checked all day")
	}
}

func TestRunRetriesAfterFailure(t *testing.T) {
	svc := &fakeService{reminders: pending()}
	rec := &recorder{err: errors.New("network down")}

	if _, err := Run(context.Background(), svc, testCfg, at(10, 0), rec.send); err == nil {
		t.Fatal("Run() should return the sending error")
	}
	if len(svc.sent["2026-10-19"]) != 0 {
		t.Fatal("a failed time must not be marked")
	}

	rec.err = nil
	if sent, err := Run(context.Background(), svc, testCfg, at(10, 15), rec.send); err != nil || !sent {
		t.Fatalf("retry: Run() = %v, %v; want sent", sent, err)
	}
}

func TestRunOffDoesNothing(t *testing.T) {
	svc := &fakeService{reminders: pending()}
	rec := &recorder{}
	if sent, err := Run(context.Background(), svc, Config{}, at(19, 0), rec.send); sent || err != nil {
		t.Fatalf("Run() without webhook = %v, %v", sent, err)
	}
}

func TestMessage(t *testing.T) {
	r := tasks.Reminders{
		Overdue: []tasks.TodoView{{Name: "Pay bill", ExpiresAt: time.Date(2026, 10, 17, 0, 0, 0, 0, time.Local)}},
		DueSoon: []tasks.TodoView{
			{Name: "Report", ExpiresAt: at(23, 59)},
			{Name: "Dentist", ExpiresAt: time.Date(2026, 10, 20, 0, 0, 0, 0, time.Local)},
			{Name: "Trip", ExpiresAt: time.Date(2026, 10, 22, 0, 0, 0, 0, time.Local)},
		},
	}

	msg := Message(r, testCfg, at(10, 0))

	for _, want := range []string{
		"<@123456789012345678>",
		"Pay bill",
		"17/10",
		"Report — " + i18n.T("notify.today"),
		"Dentist — " + i18n.T("notify.tomorrow"),
		"22/10",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message misses %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, i18n.T("notify.daily")) {
		t.Errorf("message has an empty daily section:\n%s", msg)
	}
}

func TestMessageFitsDiscordLimit(t *testing.T) {
	var r tasks.Reminders
	for range 200 {
		r.DueSoon = append(r.DueSoon, tasks.TodoView{Name: strings.Repeat("x", 30), ExpiresAt: at(23, 0)})
	}

	msg := Message(r, testCfg, at(10, 0))
	if n := len([]rune(msg)); n > maxMessage {
		t.Fatalf("message has %d characters, limit is %d", n, maxMessage)
	}
	if !strings.Contains(msg, "…") {
		t.Errorf("a cut message must say how many were left out:\n%s", msg[len(msg)-80:])
	}
}

func TestDiscordPayload(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := testCfg
	cfg.Webhook = server.URL
	if err := Discord(cfg)(context.Background(), "hello"); err != nil {
		t.Fatalf("send error = %v", err)
	}

	if got["content"] != "hello" {
		t.Errorf("content = %v", got["content"])
	}
	users := got["allowed_mentions"].(map[string]any)["users"].([]any)
	if len(users) != 1 || users[0] != cfg.Mention {
		t.Errorf("allowed_mentions = %v, want only the configured user", got["allowed_mentions"])
	}
}

func TestDiscordErrorsHideTheWebhookSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message": "Unknown Webhook"}`, http.StatusNotFound)
	}))
	cfg := Config{Webhook: server.URL + "/api/webhooks/1/secret-token"}

	err := Discord(cfg)(context.Background(), "hello")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %v, want the 404", err)
	}

	server.Close()
	err = Discord(cfg)(context.Background(), "hello")
	if err == nil || strings.Contains(err.Error(), "secret-token") {
		t.Errorf("error = %v; must fail without showing the URL", err)
	}
}

func TestConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := LoadConfig()
	if err != nil || cfg.Enabled() {
		t.Fatalf("LoadConfig() without a file = %+v, %v", cfg, err)
	}
	if cfg.DaysAhead() != 3 || !slices.Equal(cfg.Slots(), []string{"10:00", "19:00"}) || cfg.DailySlot() != "19:00" {
		t.Errorf("defaults: days %d, times %v, daily %s", cfg.DaysAhead(), cfg.Slots(), cfg.DailySlot())
	}

	zero := 0
	want := Config{Webhook: "https://discord.com/api/webhooks/1/abc", Days: &zero, Times: []string{"08:30"}}
	if err := SaveConfig(want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfig()
	if err != nil || got.DaysAhead() != 0 || got.DailySlot() != "08:30" {
		t.Errorf("LoadConfig() = %+v, %v; want days 0 and the daily slot moved to 08:30", got, err)
	}

	if times, err := ParseTimes("19:00, 9:05,19:00"); err != nil || !slices.Equal(times, []string{"09:05", "19:00"}) {
		t.Errorf("ParseTimes() = %v, %v", times, err)
	}
	if _, err := ParseTimes("25:00"); err == nil {
		t.Error("ParseTimes() accepted 25:00")
	}

	for url, ok := range map[string]bool{
		"https://discord.com/api/webhooks/123/abc-DEF_9":    true,
		"https://discordapp.com/api/webhooks/123/abc":       true,
		"https://ptb.discord.com/api/webhooks/123/abc":      true,
		"http://discord.com/api/webhooks/123/abc":           false,
		"https://evil.example/discord.com/api/webhooks/1/a": false,
	} {
		if ValidWebhook(url) != ok {
			t.Errorf("ValidWebhook(%q) = %v", url, !ok)
		}
	}
}
