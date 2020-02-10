package version

import (
	"bytes"
	"fmt"
	"runtime"
)

// GitCommit returns the git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version returns the main version number that is being run at the moment.
var Version = "0.1.0"

// BuildDate returns the date the binary was built
var BuildDate = ""

// GoVersion returns the version of the go runtime used to compile the binary
var GoVersion = runtime.Version()

// OsArch returns the os and arch used to build the binary
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

//Get Gets Version ¯\_(ツ)_/¯
func Get() string {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintln(buf, "Build Date:", BuildDate)
	_, _ = fmt.Fprintln(buf, "Git Commit:", GitCommit)
	_, _ = fmt.Fprintln(buf, "Version:", Version)
	_, _ = fmt.Fprintln(buf, "Go Version:", GoVersion)
	_, _ = fmt.Fprintln(buf, "OS / Arch:", OsArch)
	return buf.String()
}

//Print Prints the Version ¯\_(ツ)_/¯
func Print() {
	fmt.Println(Get())
}
