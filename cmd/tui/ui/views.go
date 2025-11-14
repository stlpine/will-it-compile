package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/stlpine/will-it-compile/pkg/models"
)

// viewEditor renders the code editor view
func (m Model) viewEditor() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("will-it-compile - Code Editor")
	b.WriteString(title + "\n\n")

	// Language selector
	langInfo := fmt.Sprintf("Language: %s (press 'l' to change)", m.language)
	b.WriteString(mutedStyle.Render(langInfo) + "\n\n")

	// Editor
	editorBox := activeEditorStyle.Render(m.editor.View())
	b.WriteString(editorBox + "\n\n")

	// Buttons
	compileBtn := activeButtonStyle.Render(" Compile (Enter) ")
	fileBtn := inactiveButtonStyle.Render(" Load File (f) ")
	historyBtn := inactiveButtonStyle.Render(" History (Tab) ")
	helpBtn := inactiveButtonStyle.Render(" Help (?) ")

	buttons := lipgloss.JoinHorizontal(lipgloss.Left, compileBtn, " ", fileBtn, " ", historyBtn, " ", helpBtn)
	b.WriteString(buttons + "\n")

	return b.String()
}

// viewHistory renders the job history view
func (m Model) viewHistory() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Compilation History")
	b.WriteString(title + "\n\n")

	if len(m.jobHistory) == 0 {
		b.WriteString(mutedStyle.Render("No jobs yet. Press Tab to go back to editor.\n"))
		return b.String()
	}

	// Job list
	for i, job := range m.jobHistory {
		var itemStyle lipgloss.Style
		var prefix string

		if i == m.historyIndex {
			itemStyle = selectedItemStyle
			prefix = "▶ "
		} else {
			itemStyle = normalItemStyle
			prefix = "  "
		}

		// Status icon
		var statusIcon string
		var statusColor lipgloss.Style
		switch job.Status {
		case models.StatusCompleted:
			if job.Result != nil && job.Result.Compiled {
				statusIcon = "✓"
				statusColor = successStyle
			} else {
				statusIcon = "✗"
				statusColor = errorStyle
			}
		case models.StatusProcessing:
			statusIcon = "●"
			statusColor = warningStyle
		case models.StatusQueued:
			statusIcon = "○"
			statusColor = mutedStyle
		case models.StatusFailed, models.StatusTimeout:
			statusIcon = "✗"
			statusColor = errorStyle
		}

		// Format job info
		timestamp := job.CreatedAt.Format("15:04:05")
		lang := string(job.Language)
		jobInfo := fmt.Sprintf("%s%s %s | %s | %s",
			prefix,
			statusColor.Render(statusIcon),
			timestamp,
			lang,
			truncate(job.ID, 8),
		)

		b.WriteString(itemStyle.Render(jobInfo) + "\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate  Enter: view details  Tab: back to editor  q: quit\n"))

	return b.String()
}

// viewJobDetail renders the job detail view
func (m Model) viewJobDetail() string {
	if m.currentJob == nil {
		return mutedStyle.Render("No job selected")
	}

	var b strings.Builder

	job := m.currentJob

	// Title
	title := titleStyle.Render(fmt.Sprintf("Job Details: %s", truncate(job.ID, 12)))
	b.WriteString(title + "\n\n")

	// Job info
	infoBox := boxStyle.Render(fmt.Sprintf(
		"Status: %s\nLanguage: %s\nCreated: %s",
		colorizeStatus(job.Status),
		job.Language,
		job.CreatedAt.Format("2006-01-02 15:04:05"),
	))
	b.WriteString(infoBox + "\n\n")

	// Result if available
	if job.Result != nil {
		result := job.Result

		// Result summary
		var resultSummary string
		if result.Success {
			if result.Compiled {
				resultSummary = successStyle.Render("✓ Compilation Successful")
			} else {
				resultSummary = errorStyle.Render("✗ Compilation Failed")
			}
		} else {
			resultSummary = errorStyle.Render("✗ Error: " + result.Error)
		}

		b.WriteString(resultSummary + "\n\n")

		// Details
		details := fmt.Sprintf(
			"Exit Code: %d\nDuration: %s",
			result.ExitCode,
			formatDuration(result.Duration),
		)
		b.WriteString(boxStyle.Render(details) + "\n\n")

		// Stdout
		if result.Stdout != "" {
			stdoutBox := boxStyle.Width(min(m.width-10, 100)).Render(
				fmt.Sprintf("STDOUT:\n%s", truncate(result.Stdout, 500)),
			)
			b.WriteString(stdoutBox + "\n\n")
		}

		// Stderr
		if result.Stderr != "" {
			stderrBox := boxStyle.Width(min(m.width-10, 100)).Render(
				fmt.Sprintf("STDERR:\n%s", truncate(result.Stderr, 500)),
			)
			b.WriteString(stderrBox + "\n\n")
		}
	} else {
		// Still processing
		processing := warningStyle.Render(fmt.Sprintf("%s Processing...", m.spinner.View()))
		b.WriteString(processing + "\n\n")
	}

	// Help
	b.WriteString(helpStyle.Render("Esc: back to editor  q: quit\n"))

	return b.String()
}

// viewFilePicker renders the file picker view
func (m Model) viewFilePicker() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Select a File")
	b.WriteString(title + "\n\n")

	// File picker
	b.WriteString(m.filePicker.View() + "\n\n")

	// Help
	b.WriteString(helpStyle.Render("↑/↓: navigate  Enter: select  Esc: cancel\n"))

	return b.String()
}

// viewHelp renders the help screen
func (m Model) viewHelp() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Help - Keyboard Shortcuts")
	b.WriteString(title + "\n\n")

	// Shortcuts
	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Enter", "Submit code for compilation (in editor)"},
		{"f", "Open file picker to load code from file"},
		{"l", "Change programming language"},
		{"Tab", "Toggle between editor and history"},
		{"↑/↓", "Navigate in history or file picker"},
		{"Enter", "View job details (in history)"},
		{"?", "Show this help screen"},
		{"Esc", "Go back to editor"},
		{"q / Ctrl+C", "Quit the application"},
	}

	for _, sc := range shortcuts {
		line := fmt.Sprintf("%s  %s",
			helpKeyStyle.Render(fmt.Sprintf("%-12s", sc.key)),
			sc.desc,
		)
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")

	// Features
	features := titleStyle.Render("Features")
	b.WriteString(features + "\n\n")

	featureList := []string{
		"• Write or paste code directly in the editor",
		"• Load code from local files (.cpp, .c, .go, .rs, .py)",
		"• Submit code to API server for compilation",
		"• View live compilation status updates",
		"• Browse job history and view detailed results",
		"• See stdout, stderr, and exit codes",
	}

	for _, feat := range featureList {
		b.WriteString(mutedStyle.Render(feat) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Esc or ? to close help\n"))

	return b.String()
}

// Helper functions

func colorizeStatus(status models.JobStatus) string {
	switch status {
	case models.StatusCompleted:
		return successStyle.Render(string(status))
	case models.StatusProcessing:
		return warningStyle.Render(string(status))
	case models.StatusQueued:
		return mutedStyle.Render(string(status))
	case models.StatusFailed, models.StatusTimeout:
		return errorStyle.Render(string(status))
	default:
		return string(status)
	}
}
