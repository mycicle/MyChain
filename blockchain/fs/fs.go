package fs

import (
	"os"
	"os/user"
	"path"
	"strings"
)

// replace tilde with users home dir
// expands embedded environment variables
// cleans the path, ex /a/b/../c -> a/c
func ExpandPath(p string) string {
	if i := strings.Index(p, ":"); i > 0 {
		return p
	}

	if i := strings.Index(p, "@"); i > 0 {
		return p
	}

	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := homeDir(); home != "" {
			p = home + p[1:]
		}
	}

	return path.Clean(os.ExpandEnv(p))
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}

	return ""
}