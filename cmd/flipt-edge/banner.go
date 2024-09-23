package main

const bannerTmpl = `
    _________       __      ______    __         
   / ____/ (_)___  / /_    / ____/___/ /___ ____ 
  / /_  / / / __ \/ __/   / __/ / __  / __ ./ _ \
 / __/ / / / /_/ / /_    / /___/ /_/ / /_/ /  __/
/_/   /_/_/ .___/\__/   /_____/\__,_/\__, /\___/ 
         /_/                        /____/       

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
