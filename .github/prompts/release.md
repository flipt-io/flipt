Extract and format the release notes for this specific release.

## Instructions

1. IMPORTANT: Read the CHANGELOG.md file from the Git tag being released ({{.Tag}}), NOT from the current branch
   - Use the command: `git show {{.Tag}}:CHANGELOG.md` to get the CHANGELOG.md content from the tagged commit
   - This ensures we get the exact changelog content that was part of the release, not any subsequent changes
2. Find the section that corresponds to version {{.Tag}} (without the 'v' prefix if present)
3. Extract ONLY the content for that specific version:
   - Start after the version heading line
   - Stop before the next version heading or end of file
   - Include all subsections (### Added, ### Fixed, ### Changed, ### Dependencies, etc.)
4. Return ONLY the extracted content, preserving the exact formatting
5. Do NOT include the version heading line itself (the line with ## [version])
6. Do NOT include any content from other versions
7. Do NOT add any additional commentary or formatting
8. If version {{.Tag}} is not found in CHANGELOG.md, return "Release notes for version {{.Tag}} not found in CHANGELOG.md"

## Version Information

Project name: {{.ProjectName}}
Git repository URL: {{.GitURL}}
{{ if eq .Tag "" }}
Previous version: {{.Version}}
Version: to be defined
{{ else }}
Previous Version: {{.PreviousTag}}
Version: {{.Tag}}
{{ end }}
{{ if .IsSnapshot }}This is a snapshot build.{{ end }}
{{ if .IsNightly }}This is a nightly build.{{ end }}
