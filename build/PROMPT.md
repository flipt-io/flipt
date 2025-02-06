Write the release notes for this release.

## Instructions:

- Do not use emojis
- Write a summary featuring the 2 or 3 most exciting or user-impacting changes. 
- Format the summary as a list of bullet points.
- Respond ONLY THE GENERATED RELEASE NOTES, NOTHING ELSE.
- Format the response as a markdown. Do not wrap the response in ```markdown``` tags.
- Do not change the original commit messages
- Group the changes in the release notes by type (ie: ### Added, ### Fixed, ### Changed)
- The only type headings that should be used are: Added, Changed, Deprecated, Removed, Fixed, and Security.
- The only type headings that should be used are: Added, Changed, Deprecated, Removed, Fixed, and Security.
- If there are changes for a type do not leave them out of the release notes.
- If there are no changes for a type, you can just leave that type out of the release notes.
- If there are no changes for a type, you can just leave that type out of the release notes.
- We use conventional commits <https://www.conventionalcommits.org/en/v1.0.0/>, so use the type of the commit to group them in the release notes.
    Example:

    ## Summary

    - Add new feature
    - Fix bug
    - Improve performance

    ## What's Changed

    ### Added

      - feat: add new feature

    ### Fixed

      - fix: fix bug

    ### Changed

      - perf: improve performance
      - chore: update dependencies
      - refactor: refactor code


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