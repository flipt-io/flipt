package main

const bannerTmpl = `
    _________       __ 
   / ____/ (_)___  / /_
  / /_  / / / __ \/ __/
 / __/ / / / /_/ / /_  
/_/   /_/_/ .___/\__/  
         /_/           

{{if .Version}}Version: {{.Version}}{{end}}
Commit: {{.Commit}}
Build Date: {{.Date}}
Go Version: {{.GoVersion}}
OS/Arch: {{.GoOS}}/{{.GoArch}}
`

type bannerOpts struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
	GoOS      string
	GoArch    string
}
