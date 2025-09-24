package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Common color palette used across CLI commands
var (
	Purple      = lipgloss.Color("#7571F9") // Primary accent color
	Green       = lipgloss.Color("#10B981") // Success states
	Amber       = lipgloss.Color("#F59E0B") // Warning states
	Red         = lipgloss.Color("#EF4444") // Error states
	MutedGray   = lipgloss.Color("#9CA3AF") // Muted text - lighter for better contrast
	LightGray   = lipgloss.Color("#F3F4F6") // Light backgrounds
	DarkGray    = lipgloss.Color("#D1D5DB") // Dark text - much lighter for readability
	Light       = lipgloss.Color("#E5E7EB") // Light text for labels and values
	White       = lipgloss.Color("#FFFFFF") // White text for buttons
	BorderGray  = lipgloss.Color("#4B5563") // Border color (darker for visibility)
	CodeBlockBg = lipgloss.Color("#1F2937") // Background for code blocks
)

// Common styles used across CLI commands
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Purple).
			Padding(1, 2).
			Align(lipgloss.Center)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedGray).
			Italic(true).
			Align(lipgloss.Center)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Amber).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	// Card container style
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderGray).
			Padding(1, 2).
			MarginBottom(1)

	// Section header style
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(White).
				Bold(true).
				MarginBottom(1)

	// Progress indicator styles
	ProgressActiveStyle = lipgloss.NewStyle().
				Foreground(Purple).
				Bold(true)

	ProgressInactiveStyle = lipgloss.NewStyle().
				Foreground(MutedGray)

	ProgressCompleteStyle = lipgloss.NewStyle().
				Foreground(Green).
				Bold(true)

	// Input and form styles
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(White). // Make labels clearer
			Bold(true).
			MarginBottom(1)

	HelperTextStyle = lipgloss.NewStyle().
			Foreground(MutedGray).
			Italic(true)

	// Button styles
	PrimaryButtonStyle = lipgloss.NewStyle().
				Foreground(White).
				Background(Purple).
				Padding(0, 3).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple)

	SecondaryButtonStyle = lipgloss.NewStyle().
				Foreground(Purple).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple).
				Padding(0, 3)

	// Simplified accent style - using purple as primary accent
	AccentStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(MutedGray)

	ValueStyle = lipgloss.NewStyle().
			Foreground(White). // Make values stand out more
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
