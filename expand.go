package plink

import (
	"log"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/carbocation/pfx"
)

// ExpandHome expands ~ to its proper path, where appropriate.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			log.Fatalln(pfx.Err(err))
		}
		path = filepath.Join(usr.HomeDir, (path)[2:])
	}

	return path
}
