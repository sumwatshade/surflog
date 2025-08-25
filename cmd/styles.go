package cmd

import "github.com/charmbracelet/lipgloss"

// Centralized styles for consistent UX across views.
var (
	appTitle       = "surflog"
	headerStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).Background(lipgloss.Color("57")).Padding(0, 1)
	tabStyle       = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("247"))
	activeTabStyle = tabStyle.Bold(true).Foreground(lipgloss.Color("51")).Background(lipgloss.Color("236"))
	contentStyle   = lipgloss.NewStyle().Padding(1, 2)
	footerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)
	dividerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	helpBoxStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("249")).Padding(0, 1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
)

func tabs(current string, width int) string {
	// Only journal/create are switchable; buoy data always visible left.
	names := []string{"journal", "create"}
	var rendered []string
	for _, n := range names {
		if n == current {
			rendered = append(rendered, activeTabStyle.Render(n))
		} else {
			rendered = append(rendered, tabStyle.Render(n))
		}
	}
	line := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	if width > 0 {
		// Ensure line doesn't overflow; truncate softly.
		line = lipgloss.NewStyle().MaxWidth(width).Render(line)
	}
	return line
}
