# Flipt CLI TUI Design System

This document outlines the design principles, patterns, and guidelines for building Terminal User Interface (TUI) components in the Flipt CLI. All interactive commands should follow these patterns to ensure a consistent and delightful user experience.

## Table of Contents

1. [Design Principles](#design-principles)
2. [Color System](#color-system)
3. [Typography and Styles](#typography-and-styles)
4. [Layout System](#layout-system)
5. [Component Patterns](#component-patterns)
6. [Wizard Pattern](#wizard-pattern)
7. [Forms and Input](#forms-and-input)
8. [Progress Indicators](#progress-indicators)
9. [Error Handling](#error-handling)
10. [Success States](#success-states)
11. [Code Examples](#code-examples)
12. [Best Practices](#best-practices)

## Design Principles

### 1. Progressive Disclosure
- Guide users through complex processes step-by-step
- Reveal information and options as needed
- Avoid overwhelming users with too many choices at once

### 2. Clear Visual Hierarchy
- Use badges, colors, and spacing to create clear information hierarchy
- Most important information should be most prominent
- Group related information together

### 3. Responsive Design
- Adapt to different terminal widths
- Gracefully handle narrow terminals (minimum 48 characters)
- Optimize for standard 80-character width

### 4. Consistent Experience
- All commands should feel like part of the same product
- Reuse patterns and components across commands
- Maintain consistent spacing and alignment

### 5. Delightful Interactions
- Provide immediate visual feedback
- Celebrate successful completions
- Make errors helpful, not frustrating

## Color System

Our color palette is defined in `styles.go` and uses the Charm bracelet lipgloss library:

```go
// Primary Brand Colors
Purple       = lipgloss.Color("#6366F1") // Primary accent
PurpleAccent = lipgloss.Color("#8B5CF6") // Accent highlight
PurpleLight  = lipgloss.Color("#A78BFA") // Light purple for subtle highlights
PurpleDark   = lipgloss.Color("#4C1D95") // Dark purple for borders

// Semantic Colors
Green      = lipgloss.Color("#22C55E") // Success states
GreenLight = lipgloss.Color("#86EFAC") // Light green for success highlights
Amber      = lipgloss.Color("#F59E0B") // Warning states
Red        = lipgloss.Color("#EF4444") // Error states

// Neutral Colors
MutedGray    = lipgloss.Color("#94A3B8") // Secondary text
SoftGray     = lipgloss.Color("#E2E8F0") // Label text
DarkGray     = lipgloss.Color("#64748B") // Darker secondary text
White        = lipgloss.Color("#F8FAFC") // Primary text on dark surfaces

// Surface Colors
Surface      = lipgloss.Color("#111827") // Card background surface
SurfaceMuted = lipgloss.Color("#1F2937") // Muted surface for nested areas
CodeBlockBg  = lipgloss.Color("#1F2937") // Background for code blocks
```

### Color Usage Guidelines

- **Purple**: Primary brand color for headers, progress bars, and primary actions
- **Green**: Success states, completed steps, active status
- **Amber**: Warnings, action required, important notices
- **Red**: Errors, invalid states, destructive actions
- **Grays**: Secondary text, labels, dividers

## Typography and Styles

### Text Styles

```go
TitleStyle        // Large, bold, purple headers
SubtitleStyle     // Secondary headers, muted gray
SectionHeaderStyle // Section headers within cards
LabelStyle        // Form labels and field names
ValueStyle        // User values and important data
HelperTextStyle   // Help text and descriptions
HintStyle         // Subtle hints and tips
ErrorTextStyle    // Error messages
```

### Badge Styles

Badges are used to categorize and highlight information:

```go
BadgeSuccessStyle // Green background - completed/active states
BadgeInfoStyle    // Purple background - informational
BadgeWarnStyle    // Amber background - warnings
BadgeErrorStyle   // Red background - errors
```

Usage example:
```go
badge.Render("ACTIVE")  // Creates a colored badge with text
```

## Layout System

### Content Width

- Standard content width: 76 characters
- Minimum content width: 48 characters
- Responsive width calculation:

```go
func availableWidth() int {
    width, _, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil || width == 0 {
        return contentWidth
    }
    
    usable := width - 4  // Account for padding
    if usable < minContentWidth {
        return max(1, usable)
    }
    if usable > contentWidth {
        return contentWidth
    }
    return usable
}
```

### Card System

Cards are the primary container for content:

```go
CardStyle         // Standard card with rounded border
InfoCardStyle     // Purple border for information
SuccessCardStyle  // Green border for success
WarningCardStyle  // Amber border for warnings
```

### Spacing

- Use `applySectionSpacing()` between major sections
- Cards have internal padding of 1 line vertically, 3 characters horizontally
- Maintain consistent spacing between elements

## Component Patterns

### Content Sections

The `contentSection` struct is the building block for structured content:

```go
type contentSection struct {
    badge       lipgloss.Style  // Badge style
    badgeText   string         // Badge text (e.g., "INFO", "SUCCESS")
    heading     string         // Section heading
    helperText  string         // Optional helper text
    configItems []string       // Key-value pairs or config items
    bulletItems []string       // Bullet list items
}
```

Usage:
```go
section := &contentSection{
    badge:      BadgeInfoStyle,
    badgeText:  "CONFIG",
    heading:    "License Configuration",
    helperText: "Your current license settings",
    configItems: []string{
        renderKeyValue("Type", "Pro Annual"),
        renderKeyValue("Status", "Active"),
    },
}
fmt.Println(CardStyle.Render(section.render()))
```

### Hero Headers

Hero headers provide context and show progress:

```go
func heroHeader(title, subtitle string) string {
    // Creates centered title with decorative border
    // Shows progress bar if in wizard mode
    // Adapts to terminal width
}
```

### Key-Value Pairs

For displaying configuration and status information:

```go
renderKeyValue("Label", ValueStyle.Render("value"))
// Output: Label: value
```

### Bullet Lists

For listing features or options:

```go
renderBulletList([]string{
    "First item",
    "Second item",
    "Third item",
})
// Output:
//   › First item
//   › Second item
//   › Third item
```

## Wizard Pattern

Multi-step processes should use the wizard pattern:

### 1. Define Steps

```go
type WizardStep int

const (
    StepWelcome WizardStep = iota
    StepConfiguration
    StepValidation
    StepComplete
)
```

### 2. Track Progress

```go
type wizard struct {
    currentStep WizardStep
    totalSteps  int
}
```

### 3. Show Progress

Progress bars automatically appear in hero headers when not on first/last step:

```go
// Thin progress bar with filled/unfilled sections
▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
```

### 4. Step Transitions

- Clear screen between steps: `fmt.Print("\033[H\033[2J")`
- Update `currentStep` before rendering
- Provide clear navigation

## Forms and Input

Using the `huh` library for forms:

### Form Creation

```go
form := huh.NewForm(groups...).WithTheme(huh.ThemeCharm())
form = form.WithProgramOptions(
    tea.WithOutput(os.Stdout),
    tea.WithAltScreen(),
    tea.WithReportFocus(),
)
```

### Input Types

1. **Text Input**
```go
huh.NewInput().
    Title(InputLabelStyle.Render("License Key")).
    Description("Enter your license key").
    Placeholder("XXXXX-XXXXX-XXXXX-XXXXX").
    Value(&licenseKey).
    EchoMode(huh.EchoModePassword)  // For sensitive data
```

2. **Select**
```go
huh.NewSelect[string]().
    Title("Select License Type").
    Options(
        huh.NewOption("Pro Monthly", "monthly"),
        huh.NewOption("Pro Annual", "annual"),
    ).
    Value(&licenseType)
```

3. **Confirm**
```go
huh.NewConfirm().
    Title("Ready to begin?").
    Description("This will take about 2 minutes").
    Value(&proceed).
    Affirmative("Yes, let's start").
    Negative("No, maybe later")
```

### Validation

```go
.Validate(func(s string) error {
    if s == "" {
        return fmt.Errorf("field is required")
    }
    return nil
})
```

## Progress Indicators

### Inline Status

For real-time status updates:

```go
renderInlineStatus(BadgeInfoStyle, "CHECKING", "Validating license...")
// Output: [CHECKING] Validating license...
```

### Progress Bars

For longer operations:

```go
type progressInfo struct {
    currentStep   int
    totalSteps    int
    progressPct   float64
}
```

## Error Handling

### Error Display Pattern

```go
// Don't duplicate error messages
fmt.Println(ErrorStyle.Render("✗ Operation failed"))
fmt.Println(LabelStyle.Render("Error: ") + ValueStyle.Render(err.Error()))
fmt.Println()
fmt.Println(HelperTextStyle.Render("Try this to fix the issue"))
return nil  // Return nil to prevent Cobra from printing error again
```

### User Interrupts

```go
func isInterruptError(err error) bool {
    return errors.Is(err, tea.ErrInterrupted) || 
           errors.Is(err, huh.ErrUserAborted)
}
```

## Success States

### Success Screens

Success screens should be celebratory and informative:

```go
type successScreenSection struct {
    badge       lipgloss.Style
    badgeText   string
    heading     string
    content     []string  // For key-value pairs
    bulletItems []string  // For bullet lists
}
```

Components of a good success screen:
1. Success message/title
2. Summary of what was accomplished
3. Next steps
4. Helpful resources
5. Thank you message

Example:
```go
sections := []successScreenSection{
    {BadgeSuccessStyle, "COMPLETE", "Setup Complete", summary, nil},
    {BadgeInfoStyle, "NEXT", "What to do next", nil, nextSteps},
    {BadgeInfoStyle, "RESOURCES", "Get Help", resources, nil},
}
```

## Code Examples

### Complete Command Structure

```go
type myCommand struct {
    // Configuration
    configFile string
    
    // Wizard state
    currentStep WizardStep
    totalSteps  int
    
    // User inputs
    userInput string
}

func (c *myCommand) run(cmd *cobra.Command, args []string) error {
    // 1. Check TTY
    if !isatty.IsTerminal(os.Stdout.Fd()) {
        return fmt.Errorf("requires interactive terminal")
    }
    
    // 2. Initialize
    c.currentStep = StepWelcome
    c.totalSteps = 4
    
    // 3. Run wizard steps
    steps := []wizardStep{
        {StepWelcome, c.runWelcomeStep},
        {StepConfig, c.runConfigStep},
        {StepValidate, c.runValidateStep},
    }
    
    for _, step := range steps {
        c.currentStep = step.step
        fmt.Print("\033[H\033[2J")  // Clear screen
        
        if err := step.runFunc(); err != nil {
            if isInterruptError(err) {
                fmt.Println(HelperTextStyle.Render("Cancelled."))
                return nil
            }
            return err
        }
    }
    
    // 4. Show success
    c.currentStep = StepComplete
    c.renderSuccessScreen()
    
    return nil
}
```

### Responsive Content Section

```go
func (c *myCommand) renderInfoSection() string {
    width := c.availableWidth()
    
    section := &contentSection{
        badge:      BadgeInfoStyle,
        badgeText:  "INFO",
        heading:    "Important Information",
        helperText: "Please read carefully",
        bulletItems: []string{
            "First important point",
            "Second important point",
            "Third important point",
        },
    }
    
    card := CardStyle.Copy().
        Width(width).
        Render(section.render())
    
    return applySectionSpacing(card)
}
```

## Best Practices

### 1. Terminal Compatibility
- Always check for TTY before running interactive commands
- Handle non-TTY gracefully with helpful error messages
- Test on different terminal emulators

### 2. Performance
- Clear screen between major transitions
- Avoid excessive redraws
- Cache width calculations when possible

### 3. Accessibility
- Provide clear contrast between text and backgrounds
- Use semantic colors consistently
- Include descriptive text for all actions

### 4. Error Recovery
- Provide actionable error messages
- Suggest fixes when possible
- Allow users to retry failed operations

### 5. Code Organization
- Keep styles in `styles.go`
- Share common components between commands
- Separate wizard steps into individual functions
- Use meaningful type and variable names

### 6. Testing
- Test with various terminal widths
- Test interruption handling (Ctrl+C)
- Test validation and error cases
- Test on different operating systems

### 7. Documentation
- Comment complex UI logic
- Document custom patterns
- Provide examples in code

## Common Patterns Reference

### Pattern: Contextual Help
Show help before user input:
```go
guidanceSection := &contentSection{
    badge:      BadgeInfoStyle,
    badgeText:  "HELP",
    heading:    "How to proceed",
    bulletItems: []string{"Step 1", "Step 2", "Step 3"},
}
fmt.Println(CardStyle.Render(guidanceSection.render()))
```

### Pattern: Status Check
Show current status with actionable next steps:
```go
if !isConfigured {
    section := &contentSection{
        badge:      BadgeWarnStyle,
        badgeText:  "ACTION REQUIRED",
        heading:    "Not Configured",
        helperText: "You need to set this up first",
    }
    // Add call-to-action
}
```

### Pattern: Comparison Table
Show options side-by-side:
```go
section := &contentSection{
    badge:      BadgeInfoStyle,
    badgeText:  "COMPARE",
    heading:    "Available Options",
    configItems: []string{
        SectionHeaderStyle.Render("Option A"),
        "  • Feature 1",
        "  • Feature 2",
        "",
        SectionHeaderStyle.Render("Option B"),
        "  • Feature 3",
        "  • Feature 4",
    },
}
```

## Resources

- [Charm Bracelet Libraries](https://charm.sh/) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Huh](https://github.com/charmbracelet/huh) - Forms
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework

---

*This design system is a living document. As we add new commands and patterns, this document should be updated to reflect current best practices.*