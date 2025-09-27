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
	PurpleLight  = lipgloss.Color("#A78BFA") // Light purple for subtle highlights
	PurpleDark   = lipgloss.Color("#4C1D95") // Dark purple for borders
	Green        = lipgloss.Color("#22C55E") // Success states
	GreenLight   = lipgloss.Color("#86EFAC") // Light green for success highlights
	Amber        = lipgloss.Color("#F59E0B") // Warning states
	Red          = lipgloss.Color("#EF4444") // Error states
	MutedGray    = lipgloss.Color("#94A3B8") // Secondary text
	SoftGray     = lipgloss.Color("#E2E8F0") // Label text
	DarkGray     = lipgloss.Color("#64748B") // Darker secondary text
	White        = lipgloss.Color("#F8FAFC") // Primary text on dark surfaces
	BorderGray   = lipgloss.Color("#6366F1") // Border highlight (matches purple theme)
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

	ContentIndentStyle = lipgloss.NewStyle().
				PaddingLeft(2)

	// Section header style
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(SoftGray).
				Bold(true).
				MarginBottom(1)

	// Legacy progress indicator styles (kept for compatibility)
	StepChipBaseStyle = lipgloss.NewStyle().
				Padding(0, 2).
				MarginRight(1).
				Border(lipgloss.RoundedBorder()).
				Align(lipgloss.Center)

	StepChipActiveStyle = StepChipBaseStyle.Copy().
				Foreground(White).
				Background(PurpleAccent).
				BorderForeground(PurpleLight).
				Bold(true)

	StepChipCompleteStyle = StepChipBaseStyle.Copy().
				Foreground(White).
				Background(Green).
				BorderForeground(GreenLight).
				Bold(true)

	StepChipInactiveStyle = StepChipBaseStyle.Copy().
				Foreground(DarkGray).
				Background(SurfaceMuted).
				BorderForeground(MutedGray)

	// Progress bar styles for continuous progress indication
	ProgressBarStyle = lipgloss.NewStyle().
				MarginTop(1).
				MarginBottom(1)

	ProgressBarTrackStyle = lipgloss.NewStyle().
				Foreground(MutedGray)

	ProgressBarFillStyle = lipgloss.NewStyle().
				Foreground(PurpleAccent)

	ProgressBarActiveStyle = lipgloss.NewStyle().
				Foreground(Purple).
				Bold(true)

	ProgressBarCompleteStyle = lipgloss.NewStyle().
					Foreground(Green)

	ProgressLabelStyle = lipgloss.NewStyle().
				Foreground(MutedGray).
				MarginLeft(2)

	ProgressCounterStyle = lipgloss.NewStyle().
				Foreground(PurpleAccent).
				Bold(true).
				MarginLeft(1)

	// Legacy progress indicator styles (kept for compatibility)
	ProgressBaseStyle = lipgloss.NewStyle().
				MarginRight(2)

	ProgressActiveStyle = ProgressBaseStyle.Copy().
				Foreground(PurpleAccent).
				Bold(true)

	ProgressCompleteStyle = ProgressBaseStyle.Copy().
				Foreground(Green)

	ProgressInactiveStyle = ProgressBaseStyle.Copy().
				Foreground(MutedGray)

	ProgressSeparatorStyle = lipgloss.NewStyle().
				Foreground(MutedGray).
				MarginLeft(1).
				MarginRight(1)

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

	// Additional accessibility styles
	FocusStyle = lipgloss.NewStyle().
			BorderForeground(PurpleLight).
			BorderStyle(lipgloss.DoubleBorder())

	ErrorTextStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	HintStyle = lipgloss.NewStyle().
			Foreground(DarkGray).
			Italic(true)
)

func applySectionSpacing(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	return SectionStyle.Render(content)
}
