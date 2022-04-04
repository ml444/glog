package util

import "strings"

func GetPackageName(f string) (string, string) {
	slashIndex := strings.LastIndex(f, "/")
	if slashIndex > 0 {
		idx := strings.Index(f[slashIndex:], ".") + slashIndex
		return f[:idx], f[idx+1:]
	}

	return f, ""
}
