package compileinfo

import (
	"fmt"
	"os"
	"runtime/debug"
)

type CompileInfo struct {
	Package    string
	GoVersion  string
	Commit     string
	CommitTime string
	Modified   bool
}

func (c CompileInfo) String() string {
	mod := ""
	if c.Modified {
		mod = " Files in the repo were modified after that commit."
	}

	return fmt.Sprintf("This %s binary was built with %s at commit %v at time %v.%s", c.Package, c.GoVersion, c.Commit, c.CommitTime, mod)
}

func Get() CompileInfo {
	out := CompileInfo{}

	z, ok := debug.ReadBuildInfo()
	if !ok {
		return out
	}

	out.GoVersion = z.GoVersion
	out.Package = z.Path
	for _, s := range z.Settings {
		switch s.Key {
		case "vcs.revision":
			out.Commit = s.Value
		case "vcs.time":
			out.CommitTime = s.Value
		case "vcs.modified":
			out.Modified = s.Value == "true"
		}
	}

	return out
}

func PrintToStdErr() {
	z := Get()
	fmt.Fprintf(os.Stderr, "%s\n", z)
}
