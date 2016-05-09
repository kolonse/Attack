// Flag
package main

import (
	"flag"
)

var Port = flag.Int("P", 8099, "listen port")

func init() {
	flag.Parse()
}
