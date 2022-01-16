package main

const bannerTmpl = `
 _____ _ _       _
|  ___| (_)_ __ | |_
| |_  | | | '_ \| __|
|  _| | | | |_) | |_
|_|   |_|_| .__/ \__|
          |_|

{{if .Version}}Version: {{.Version}}{{end}}
Commit: {{.Commit}}
Build Date: {{.Date}}
Go Version: {{.GoVersion}}
`

type bannerOpts struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
}
