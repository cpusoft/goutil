package urlutil

import (
	"net/url"
	"strings"
)

func HostAndPath(urlStr string) (string, error) {

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	pos := strings.LastIndex(u.Path, "/")
	return (u.Host + string(u.Path[:pos+1])), nil
}
