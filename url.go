package server

import (
	"bytes"

	exbytes "github.com/arex0/go-ext/bytes"
)

// ParseSelector split a url with first '/', return first part.
func ParseSelector(url string) string {
	if len(url) < 2 {
		return ""
	}
	u := exbytes.FromStringUnsafe(url)
	if i := bytes.IndexByte(u[1:], '/');i != -1 {
		return url[1 :i+1]
	}
	return url[1:]
}

// ParseURL split a url with  '/'
func ParseURL(url string) []string {
	var s []string
	u := exbytes.FromStringUnsafe(url)
	for i := 1;i < len(u);{
		if j := bytes.IndexByte(u[1:], '/');j != -1 {
			s = append(s, url[i:i+j])
			i += j + 1
		} else {
			s = append(s, url[i:])
			break
		}
	}
	return s
}
