Write the release notes for this release.

## Instructions

- do not use emojis
- write short summary featuring the 2 or 3 most exciting or user-impacting changes, if it makes sense to do so.
- keep dependency updates brief unless they're security-related or major version changes
- respond ONLY THE GENERATED RELEASE NOTES, NOTHING ELSE.
- format the response as a markdown
- do not change the original commit messages

## Some details about this release

Project name: {{.ProjectName}}
Git repository URL: {{.GitURL}}
{{ if eq .Tag "" }}
Previous version: {{.Version}}
Version: to be defined
{{ else }}
Previous Version: {{.PreviousTag}}
Version: {{.CurrentTag}}
{{ end }}
{{ if .IsSnapshot }}This is a snapshot build.{{ end }}
{{ if .IsNightly }}This is a nightly build.{{ end }}
{{ with .TagSubject }}Tag subject: {{ . }}{{ end }}
{{ with .TagContents }}Tag content: {{ . }}{{ end }}

## Release notes

{{ .ReleaseNotes }}
