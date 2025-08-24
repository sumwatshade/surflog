package buoy

import "github.com/charmbracelet/lipgloss"

var buoyTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
var buoyInfoStyle = lipgloss.NewStyle().Faint(true)

func View(data *Data) string {
	if data == nil {
		return buoyTitleStyle.Render("Buoy Data") + "\n" + buoyInfoStyle.Render("No buoy configured yet. Configure in $HOME/.surflog.yaml")
	}
	return buoyTitleStyle.Render("Buoy Data") + "\n" + "Buoy ID: " + data.id
}
