package colors

import "github.com/charmbracelet/lipgloss"

const (
	Error = lipgloss.Color("#ff5555")
	Logo  = lipgloss.Color("#ffffff")

	Muted     = lipgloss.Color("#777777")
	Highlight = lipgloss.Color("#f1fa8c")
	Success   = lipgloss.Color("#50fa7b")
	Frozen    = lipgloss.Color("#8be9fd")
	Modifier  = lipgloss.Color("#ffb86c")
	Multiply  = lipgloss.Color("#ff79c6")
	Action    = lipgloss.Color("#bd93f9")
)

// Players are handed out so everyone at the table has a different color.
var Players = []lipgloss.Color{
	"#3B82F6", // Blue
	"#10B981", // Green
	"#FACC15", // Yellow
	"#8B5CF6", // Purple
	"#06B6D4", // Cyan
	"#F97316", // Orange
}

// Numbers colors each number card value from 0 to 12.
var Numbers = []lipgloss.Color{
	"#d0d0d0", "#c8b6a6", "#f4a261", "#e9c46a", "#a7c957", "#52b788",
	"#2a9d8f", "#48cae4", "#4895ef", "#9d8df1", "#c77dff", "#f15bb5", "#ef476f",
}
