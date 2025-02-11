package main

import "strings"

func getParentDir(path string) string {
	n := strings.Split(path, "/")
	if len(n) < 2 {
		return ""
	}
	return n[len(n)-2]
}
