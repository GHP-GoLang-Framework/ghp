package codegen

import (
	"fmt"
	"strings"
)

// writeLine emits a `//line file:N` directive on its own line, which Go
// requires to start at column 1. It writes nothing when file is empty or
// the node has no known line, keeping directive-free output byte-identical.
func writeLine(b *strings.Builder, file string, line int) {
	if file == "" || line <= 0 {
		return
	}
	fmt.Fprintf(b, "//line %s:%d\n", file, line)
}
