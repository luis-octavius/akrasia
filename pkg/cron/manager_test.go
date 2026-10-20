package cron

import "testing"

func TestWithoutJobKeepsOtherJobs(t *testing.T) {
	m := NewManager()
	crontab := "MAILTO=me\n0 3 * * * backup.sh\n" +
		m.buildCronEntry("akrasia-notify", "*/15 * * * *", "akrasia notify", "Akrasia reminders") + "\n" +
		m.buildCronEntry("other", "0 * * * *", "other.sh", "") + "\n"

	got := withoutJob(crontab, "akrasia-notify")
	want := "MAILTO=me\n0 3 * * * backup.sh\n" +
		m.buildCronEntry("other", "0 * * * *", "other.sh", "") + "\n"
	if got != want {
		t.Errorf("withoutJob() =\n%q\nwant\n%q", got, want)
	}

	if m.jobExists(got, "akrasia-notify") || !m.jobExists(got, "other") {
		t.Error("wrong job removed")
	}
}
