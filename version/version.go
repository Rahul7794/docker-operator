package version

import (
	"fmt"
	"runtime"
)

// GitCommit : The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version : The main version number that is being run at the moment.
const Version = "0.1.0"

// BuildDate : Current Build Date
var BuildDate = ""

// GoVersion : Current GoLang Version
var GoVersion = runtime.Version()

// OsArch : Os Architecture
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
