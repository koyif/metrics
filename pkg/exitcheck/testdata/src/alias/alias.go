package alias

import (
	mylog "log"
	myos "os"
)

// Test that analyzer correctly detects log.Fatal and os.Exit
// even when imported with different aliases
func BadLogFatalAlias() {
	mylog.Fatal("bad") // want "log.Fatal should only be called in main function of main package"
}

func BadOsExitAlias() {
	myos.Exit(1) // want "os.Exit should only be called in main function of main package"
}
