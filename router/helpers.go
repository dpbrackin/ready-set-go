package router

import "strings"

func parsePattern(pattern string) (method, host, path string) {
	parts := strings.Fields(pattern)

	if len(parts) == 2 {
		method = parts[0]
		pattern = parts[1]
	} else {
		method = ""
	}

	urlParts := strings.SplitN(pattern, "/", 2)

	host = urlParts[0]

	if len(urlParts) == 2 {
		path = "/" + urlParts[1]
	} else {
		path = "/"
	}

	return
}
