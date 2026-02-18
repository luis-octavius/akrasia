package cron

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/robfig/cron/v3"
)

type Manager struct {
	parser cron.Parser
}

func NewManager() *Manager {
	return &Manager{
		parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// ValidateSchedule checks if a given schedule is valid
func (m *Manager) ValidateSchedule(schedule string) error {
	_, err := m.parser.Parse(schedule)
	return err
}

// AddJob adds a new cron job to the system crontab
func (m *Manager) AddJob(name, schedule, command, user, comment string) error {
	// get current crontab
	// Crontab manages the crontab user table files
	currentCron, err := m.getCurrentCrontab(user)
	if err != nil {
		return fmt.Errorf("failed to get current crontab: %v", err)
	}

	if m.jobExists(currentCron, name) {
		return fmt.Errorf("job with name '%s' already exists", name)
	}

	newEntry := m.buildCronEntry(name, schedule, command, comment)

	newCrontab := currentCron
	if newCrontab != "" && !strings.HasSuffix(newCrontab, "\n") {
		newCrontab += "\n"
	}
	newCrontab += newEntry + "\n"

	return m.writeCrontab(newCrontab)
}

// buildCronEntry creates a formatted cron entry
func (m *Manager) buildCronEntry(name, schedule, command, comment string) string {
	var builder strings.Builder

	if comment != "" {
		builder.WriteString(fmt.Sprintf("# %s\n", comment))
	}

	// add marker for the job name to easily identify
	builder.WriteString(fmt.Sprintf("# JOB_NAME=%s\n", name))
	builder.WriteString(fmt.Sprintf("%s %s\n", schedule, command))

	return builder.String()
}

func (m *Manager) getCurrentCrontab(user string) (string, error) {
	var cmd *exec.Cmd

	if user != "" {
		cmd = exec.Command("crontab", "-u", user, "-l")
	} else {
		cmd = exec.Command("crontab", "-l")
	}

	output, err := cmd.Output()
	if err != nil {
		// crontab might not exist
		return "", nil
	}

	return string(output), nil
}

// writeCrontab writes the crontab content
func (m *Manager) writeCrontab(content string) error {
	tmpFile, err := os.CreateTemp("", "crontab-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	var cmd *exec.Cmd
	cmd = exec.Command("crontab", tmpFile.Name())

	if err := cmd.Run(); err != nil {
		fmt.Println(cmd.Output())
		return fmt.Errorf("failed to install crontab: %v", err)
	}

	return nil
}

// jobExists checks if a job with the given name already exists
func (m *Manager) jobExists(crontab, name string) bool {
	scanner := bufio.NewScanner(strings.NewReader(crontab))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, fmt.Sprintf("JOB_NAME=%s", name)) {
			return true
		}
	}
	return false
}

func (m *Manager) ListJobs(user string) ([]map[string]string, error) {
	crontab, err := m.getCurrentCrontab(user)
	if err != nil {
		return nil, err
	}

	var jobs []map[string]string
	scanner := bufio.NewScanner(strings.NewReader(crontab))

	var currentJob map[string]string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "# JOB_NAME=") {
			// start of a new job
			if currentJob != nil {
				jobs = append(jobs, currentJob)
			}
			currentJob = make(map[string]string)
			currentJob["name"] = strings.TrimPrefix(line, "# JOB_NAME=")
		} else if strings.HasPrefix(line, "#") && currentJob != nil {
			// job comment
			currentJob["comment"] = strings.TrimPrefix(line, "# ")
		} else if line != "" && !strings.HasPrefix(line, "#") && currentJob != nil {
			// the actual cron job line
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				currentJob["schedule"] = strings.Join(parts[:5], " ")
				currentJob["command"] = strings.Join(parts[5:], " ")
			}
		}
	}

	// sees if there is still a job
	if currentJob != nil {
		jobs = append(jobs, currentJob)
	}

	return jobs, nil
}
