// +build !go1.13

package storage

import "flag"

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()
}
