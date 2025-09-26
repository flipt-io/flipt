package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const contentWidth = 76

// Common color palette used across CLI commands
var (
	Purple       = lipgloss.Color("#6366F1") // Primary accent
	PurpleAccent = lipgloss.Color("#8B5CF6") // Accent highlight
	Green        = lipgloss.Color("#22C55E") // Success states
	Amber        = lipgloss.Color("#F59E0B") // Warning states
	Red          = lipgloss.Color("#EF4444") // Error states
	MutedGray    = lipgloss.Color("#94A3B8") // Secondary text
	SoftGray     = lipgloss.Color("#E2E8F0") // Label text
	White        = lipgloss.Color("#F8FAFC") // Primary text on dark surfaces
	BorderGray   = lipgloss.Color("#4338CA") // Border highlight
	Surface      = lipgloss.Color("#111827") // Card background surface
	SurfaceMuted = lipgloss.Color("#1F2937") // Muted surface for nested areas
	CodeBlockBg  = lipgloss.Color("#1F2937") // Background for code blocks
)

// Common styles used across CLI commands
var (
	TitleStyle = lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Left).
			Bold(true).
			Foreground(PurpleAccent)

	SubtitleStyle = lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Left).
			Foreground(MutedGray).
			MarginBottom(1)

	DividerStyle = lipgloss.NewStyle().
			Foreground(BorderGray)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Amber).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	BadgeBaseStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			MarginRight(1).
			Align(lipgloss.Center).
			Foreground(White)

	BadgeSuccessStyle = BadgeBaseStyle.Copy().
				Background(Green)

	BadgeInfoStyle = BadgeBaseStyle.Copy().
			Background(PurpleAccent)

	BadgeWarnStyle = BadgeBaseStyle.Copy().
			Background(Amber)

	BadgeErrorStyle = BadgeBaseStyle.Copy().
			Background(Red)

	// Card container style
	CardStyle = lipgloss.NewStyle().
			Width(contentWidth).
			Padding(0, 2).
			MarginTop(0).
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderGray)

	// Section header style
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(SoftGray).
				Bold(true).
				MarginBottom(1)

	// Progress indicator styles
	StepChipBaseStyle = lipgloss.NewStyle().
				Padding(0, 1).
				MarginRight(2)

	StepChipActiveStyle = StepChipBaseStyle.Copy().
				Foreground(White).
				Background(PurpleAccent).
				Bold(true)

	StepChipCompleteStyle = StepChipBaseStyle.Copy().
				Foreground(Green).
				Background(SurfaceMuted)

	StepChipInactiveStyle = StepChipBaseStyle.Copy().
				Foreground(MutedGray).
				Background(SurfaceMuted)

	// Input and form styles
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(SoftGray).
			Bold(true).
			MarginBottom(1)

	HelperTextStyle = lipgloss.NewStyle().
			Foreground(MutedGray)

	// Button styles
	PrimaryButtonStyle = lipgloss.NewStyle().
				Foreground(White).
				Background(Purple).
				Padding(0, 3).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PurpleAccent)

	SecondaryButtonStyle = lipgloss.NewStyle().
				Foreground(Purple).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PurpleAccent).
				Padding(0, 3)

	// Simplified accent style - using purple as primary accent
	AccentStyle = lipgloss.NewStyle().
			Foreground(PurpleAccent).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(MutedGray)

	ValueStyle = lipgloss.NewStyle().
			Foreground(White).
			Bold(true)

	SectionStyle = lipgloss.NewStyle().
			MarginTop(1)

	ConfigItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// Keyboard help style
	KeyboardHelpStyle = lipgloss.NewStyle().
				Foreground(MutedGray).
				Align(lipgloss.Center).
				MarginTop(1)
)

func applySectionSpacing(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	return SectionStyle.Render(content)
}
