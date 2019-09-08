// +build go1.13

package storage

import (
	"flag"
	"testing"
)

func init() {
	testing.Init()
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()
}
