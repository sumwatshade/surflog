package cmd

import "github.com/charmbracelet/lipgloss"

// Centralized styles for consistent UX across views.
// Ocean palette
// Deep Blue: 25, Teal: 30/36, Cyan accents: 44/51, Soft Grey: 243-247, Dark Grey: 238, Light Foam: 159
var (
	appTitle       = "surflog"
	headerStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Background(lipgloss.Color("24")).Padding(0, 1)
	tabStyle       = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("245"))
	activeTabStyle = tabStyle.Bold(true).Foreground(lipgloss.Color("159")).Background(lipgloss.Color("24"))
	contentStyle   = lipgloss.NewStyle().Padding(1, 2)
	footerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Padding(0, 1)
	dividerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("24"))
	helpBoxStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Padding(0, 1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("24"))
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
