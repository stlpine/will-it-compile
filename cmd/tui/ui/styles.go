package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors.
	primaryColor   = lipgloss.Color("39")  // Blue
	successColor   = lipgloss.Color("42")  // Green
	errorColor     = lipgloss.Color("196") // Red
	warningColor   = lipgloss.Color("220") // Yellow
	mutedColor     = lipgloss.Color("241") // Gray
	highlightColor = lipgloss.Color("205") // Pink

	// Base styles.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle(). //nolint:unused // reserved for future TUI features
			Foreground(mutedColor).
			Italic(true)

		// Box styles.
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	activeBoxStyle = lipgloss.NewStyle(). //nolint:unused // reserved for future TUI features
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlightColor).
			Padding(1, 2)

		// Status styles.
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

		// Button styles.
	activeButtonStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(lipgloss.Color("0")).
				Padding(0, 2).
				Bold(true)

	inactiveButtonStyle = lipgloss.NewStyle().
				Background(mutedColor).
				Foreground(lipgloss.Color("0")).
				Padding(0, 2)

		// Tab styles.
	activeTabStyle = lipgloss.NewStyle(). //nolint:unused // reserved for future TUI features
			Border(lipgloss.Border{
			Top:    "─",
			Bottom: " ",
			Left:   "│",
			Right:  "│",
		}, true).
		BorderForeground(highlightColor).
		Padding(0, 1).
		Bold(true)

	inactiveTabStyle = lipgloss.NewStyle(). //nolint:unused // reserved for future TUI features
				Border(lipgloss.Border{
			Top:    "─",
			Bottom: "─",
			Left:   "│",
			Right:  "│",
		}, true).
		BorderForeground(mutedColor).
		Padding(0, 1)

		// Code editor styles.
	editorStyle = lipgloss.NewStyle(). //nolint:unused // reserved for future TUI features
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1).
			Width(80)

	activeEditorStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(highlightColor).
				Padding(1).
				Width(80)

		// List styles.
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

		// Help styles.
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

		// Status bar.
	statusBarStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1)

	statusBarErrorStyle = lipgloss.NewStyle().
				Background(errorColor).
				Foreground(lipgloss.Color("0")).
				Padding(0, 1)

	statusBarSuccessStyle = lipgloss.NewStyle().
				Background(successColor).
				Foreground(lipgloss.Color("0")).
				Padding(0, 1)
)
