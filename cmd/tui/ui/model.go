package ui

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stlpine/will-it-compile/cmd/tui/client"
	"github.com/stlpine/will-it-compile/pkg/models"
)

// ViewState represents the current view.
type ViewState int

const (
	ViewEditor ViewState = iota
	ViewHistory
	ViewJobDetail
	ViewFilePicker
	ViewHelp
)

// JobInfo combines job metadata with its result.
type JobInfo struct {
	ID        string
	Language  models.Language
	Status    models.JobStatus
	Result    *models.CompilationResult
	CreatedAt time.Time
}

// Model is the main TUI model.
type Model struct {
	client *client.Client
	apiURL string

	// Current state
	state  ViewState
	width  int
	height int

	// Components
	editor     textarea.Model
	spinner    spinner.Model
	filePicker filepicker.Model

	// Data
	sourceCode   string
	language     models.Language
	environments []models.EnvironmentSpec

	// Job management
	currentJob   *JobInfo
	jobHistory   []JobInfo
	historyIndex int
	isCompiling  bool

	// Status
	statusMsg string
	errorMsg  string

	// Config
	autoRefresh bool
}

// NewModel creates a new TUI model.
func NewModel(apiURL string) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter your code here or press 'f' to load from file..."
	ta.Focus()
	ta.CharLimit = 1024 * 1024 // 1MB
	ta.SetWidth(80)
	ta.SetHeight(20)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(primaryColor)

	fp := filepicker.New()
	fp.AllowedTypes = []string{".cpp", ".cc", ".cxx", ".c++", ".c", ".go", ".rs"}
	fp.Height = 15

	return Model{
		client:       client.NewClient(apiURL),
		apiURL:       apiURL,
		state:        ViewEditor,
		editor:       ta,
		spinner:      sp,
		filePicker:   fp,
		language:     models.LanguageCpp, // Default
		jobHistory:   []JobInfo{},
		historyIndex: 0,
		autoRefresh:  true,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
		m.checkHealth(),
		m.fetchEnvironments(),
	)
}

// Messages.
type (
	healthCheckMsg struct {
		err error
	}

	environmentsMsg struct {
		envs []models.EnvironmentSpec
		err  error
	}

	compileStartMsg struct{}

	compileResultMsg struct {
		job *models.CompilationJob
		err error
	}

	jobUpdateMsg struct {
		status *client.JobStatus
		err    error
	}

	fileSelectedMsg struct {
		path    string
		content string
		err     error
	}
)

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == ViewEditor || m.state == ViewHistory {
				return m, tea.Quit
			}
			// For other views, go back to editor
			m.state = ViewEditor
			m.errorMsg = ""
			return m, nil

		case "esc":
			if m.state != ViewEditor {
				m.state = ViewEditor
				m.errorMsg = ""
			}
			return m, nil

		case "?":
			m.state = ViewHelp
			return m, nil

		case "tab":
			// Toggle between editor and history
			switch m.state {
			case ViewEditor:
				m.state = ViewHistory
			case ViewHistory:
				m.state = ViewEditor
			}
			return m, nil
		}

		// View-specific keys
		switch m.state {
		case ViewEditor:
			return m.handleEditorKeys(msg)
		case ViewHistory:
			return m.handleHistoryKeys(msg)
		case ViewJobDetail:
			return m.handleJobDetailKeys(msg)
		case ViewFilePicker:
			return m.handleFilePickerKeys(msg)
		case ViewHelp:
			return m.handleHelpKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.editor.SetWidth(min(msg.Width-10, 100))
		m.editor.SetHeight(min(msg.Height-15, 25))

	case healthCheckMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("API server unreachable: %v", msg.err)
		} else {
			m.statusMsg = "Connected to API server"
		}

	case environmentsMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to fetch environments: %v", msg.err)
		} else {
			m.environments = msg.envs
		}

	case compileStartMsg:
		m.isCompiling = true
		m.statusMsg = "Compiling..."
		m.errorMsg = ""

	case compileResultMsg:
		m.isCompiling = false
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Compilation failed: %v", msg.err)
		} else {
			// Create JobInfo from the compilation job
			jobInfo := &JobInfo{
				ID:        msg.job.ID,
				Language:  msg.job.Request.Language,
				Status:    msg.job.Status,
				Result:    nil, // Will be populated when we poll
				CreatedAt: msg.job.CreatedAt,
			}
			m.currentJob = jobInfo
			m.jobHistory = append([]JobInfo{*jobInfo}, m.jobHistory...)
			m.state = ViewJobDetail

			// Start polling if job is still processing
			if msg.job.Status == models.StatusQueued || msg.job.Status == models.StatusProcessing {
				return m, m.pollJob(msg.job.ID)
			}
		}

	case jobUpdateMsg:
		if msg.err == nil && msg.status != nil {
			// Update current job
			if m.currentJob != nil && m.currentJob.ID == msg.status.JobID {
				m.currentJob.Status = msg.status.Status
				m.currentJob.Result = msg.status.Result
			}

			// Update in history
			for i, job := range m.jobHistory {
				if job.ID == msg.status.JobID {
					m.jobHistory[i].Status = msg.status.Status
					m.jobHistory[i].Result = msg.status.Result
					break
				}
			}

			// Continue polling if still processing
			if msg.status.Status == models.StatusQueued || msg.status.Status == models.StatusProcessing {
				return m, m.pollJob(msg.status.JobID)
			}
		}

	case fileSelectedMsg:
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to load file: %v", msg.err)
		} else {
			m.editor.SetValue(msg.content)
			m.statusMsg = "Loaded file: " + msg.path
		}
		m.state = ViewEditor

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update components based on current state
	switch m.state {
	case ViewEditor:
		m.editor, cmd = m.editor.Update(msg)
		cmds = append(cmds, cmd)
	case ViewFilePicker:
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI.
func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var content string

	switch m.state {
	case ViewEditor:
		content = m.viewEditor()
	case ViewHistory:
		content = m.viewHistory()
	case ViewJobDetail:
		content = m.viewJobDetail()
	case ViewFilePicker:
		content = m.viewFilePicker()
	case ViewHelp:
		content = m.viewHelp()
	}

	// Status bar
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

// Helper commands

func (m Model) checkHealth() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := m.client.HealthCheck(ctx)
		return healthCheckMsg{err: err}
	}
}

func (m Model) fetchEnvironments() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		envs, err := m.client.GetEnvironments(ctx)
		return environmentsMsg{envs: envs, err: err}
	}
}

func (m Model) compileCode() tea.Cmd {
	return func() tea.Msg {
		return compileStartMsg{}
	}
}

func (m Model) submitCompilation() tea.Cmd {
	code := m.editor.Value()
	lang := m.language

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		req := models.CompilationRequest{
			Code:     base64.StdEncoding.EncodeToString([]byte(code)),
			Language: lang,
		}

		job, err := m.client.SubmitCompilation(ctx, req)
		return compileResultMsg{job: job, err: err}
	}
}

func (m Model) pollJob(jobID string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond) // Poll every 500ms

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, err := m.client.GetJob(ctx, jobID)
		return jobUpdateMsg{status: status, err: err}
	}
}

func (m Model) loadFile(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return fileSelectedMsg{path: path, err: err}
		}
		return fileSelectedMsg{path: path, content: string(content)}
	}
}

func (m Model) renderStatusBar() string {
	left := fmt.Sprintf(" API: %s ", m.apiURL)

	var right string
	if m.errorMsg != "" {
		right = fmt.Sprintf(" ERROR: %s ", m.errorMsg)
		bar := statusBarErrorStyle.Render(left) + statusBarErrorStyle.Render(right)
		return statusBarErrorStyle.Width(m.width).Render(bar)
	} else if m.isCompiling {
		right = fmt.Sprintf(" %s Compiling... ", m.spinner.View())
		bar := statusBarStyle.Render(left) + statusBarStyle.Render(right)
		return statusBarStyle.Width(m.width).Render(bar)
	} else if m.statusMsg != "" {
		right = fmt.Sprintf(" %s ", m.statusMsg)
		bar := statusBarSuccessStyle.Render(left) + statusBarSuccessStyle.Render(right)
		return statusBarSuccessStyle.Width(m.width).Render(bar)
	}

	right = " Ready "
	bar := statusBarStyle.Render(left) + statusBarStyle.Render(right)
	return statusBarStyle.Width(m.width).Render(bar)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
