Write the release notes for this release.

## Instructions:

- do not use emojis
- write short summary featuring the 2 or 3 most exciting or user-impacting changes, if it makes sense to do so.
- we use conventional commits (https://www.conventionalcommits.org/en/v1.0.0/), so you can use the type of the commit to group them in the release notes.
    examples:
        - feat: add new feature
        - fix: fix bug
        - perf: improve performance
        - chore: update dependencies
        - refactor: refactor code
- if there are no changes for a type, you can just leave that type out of the release notes.
- respond ONLY THE GENERATED RELEASE NOTES, NOTHING ELSE.
- format the response as a markdown
- do not change the original commit messages

## Some details about this release:

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

## Release notes:

{{ .ReleaseNotes }}
