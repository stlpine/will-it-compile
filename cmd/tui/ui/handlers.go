package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stlpine/will-it-compile/pkg/models"
)

// handleEditorKeys handles keyboard input in the editor view.
func (m Model) handleEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.String() {
	case "enter":
		if !m.isCompiling {
			// Submit compilation
			return m, tea.Batch(
				m.compileCode(),
				m.submitCompilation(),
			)
		}

	case "f":
		// Open file picker
		m.state = ViewFilePicker
		// Initialize file picker to current directory
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(nil)
		return m, cmd

	case "l":
		// Cycle through languages
		m.language = m.cycleLanguage()
		m.statusMsg = "Language: " + string(m.language)
		return m, nil

	case "ctrl+l":
		// Clear editor
		m.editor.Reset()
		m.statusMsg = "Editor cleared"
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// handleHistoryKeys handles keyboard input in the history view.
func (m Model) handleHistoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.historyIndex > 0 {
			m.historyIndex--
		}

	case "down", "j":
		if m.historyIndex < len(m.jobHistory)-1 {
			m.historyIndex++
		}

	case "enter":
		if len(m.jobHistory) > 0 {
			m.currentJob = &m.jobHistory[m.historyIndex]
			m.state = ViewJobDetail

			// Start polling if job is still processing
			if m.currentJob.Status == models.StatusQueued || m.currentJob.Status == models.StatusProcessing {
				return m, m.pollJob(m.currentJob.ID)
			}
		}

	case "d":
		// Delete job from history
		if len(m.jobHistory) > 0 {
			m.jobHistory = append(m.jobHistory[:m.historyIndex], m.jobHistory[m.historyIndex+1:]...)
			if m.historyIndex >= len(m.jobHistory) && m.historyIndex > 0 {
				m.historyIndex--
			}
			m.statusMsg = "Job removed from history"
		}

	case "c":
		// Clear history
		m.jobHistory = []JobInfo{}
		m.historyIndex = 0
		m.statusMsg = "History cleared"
	}

	return m, nil
}

// handleJobDetailKeys handles keyboard input in the job detail view.
func (m Model) handleJobDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		// Refresh job status
		if m.currentJob != nil {
			return m, m.pollJob(m.currentJob.ID)
		}

	case "backspace":
		// Go back to history
		m.state = ViewHistory
	}

	return m, nil
}

// handleFilePickerKeys handles keyboard input in the file picker view.
func (m Model) handleFilePickerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if file was selected
	if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
		return m, m.loadFile(path)
	}

	// Check if disabled (file selected but not allowed type)
	if didSelect, _ := m.filePicker.DidSelectDisabledFile(msg); didSelect {
		m.errorMsg = "File type not supported"
		m.state = ViewEditor
		return m, nil
	}

	return m, nil
}

// handleHelpKeys handles keyboard input in the help view.
func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Any key exits help
	m.state = ViewEditor
	return m, nil
}

// cycleLanguage cycles to the next available language.
func (m Model) cycleLanguage() models.Language {
	languages := []models.Language{
		models.LanguageCpp,
		models.LanguageC,
		models.LanguageGo,
		models.LanguageRust,
	}

	// Find current index
	currentIndex := 0
	for i, lang := range languages {
		if lang == m.language {
			currentIndex = i
			break
		}
	}

	// Move to next
	nextIndex := (currentIndex + 1) % len(languages)
	return languages[nextIndex]
}
