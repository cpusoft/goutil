package urlutil

import (
	"net/url"
	"strings"
)

// http://server:port/aa/bb/cc.html --> server/aa/bb/
func HostAndPath(urlStr string) (string, error) {

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	pos := strings.LastIndex(u.Path, "/")
	host := u.Host
	// if have port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return (host + string(u.Path[:pos+1])), nil
}

// http://server:port/aa/bb/cc.html --> server/aa/bb/cc.html
func HostAndPathFile(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	host := u.Host
	// if have port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return (host + u.Path), nil
}

func IsUrl(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}
