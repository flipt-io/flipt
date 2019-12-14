package main

const bannerTmpl = `
______ _         _
|  ___| (_)_ __ | |_
| |_  | | | '_ \| __|
|  _| | | | |_) | |_
|_|   |_|_| .__/ \__|
         |_|

Version: {{.Version}}
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
