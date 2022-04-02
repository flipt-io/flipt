//go:build assets
// +build assets

package swagger

import "embed"

//go:embed flipt.swagger.json index.html
var Docs embed.FS
