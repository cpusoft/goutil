package osutil

import (
	"net/url"
	"strings"
)

// Deprecated: using urlutil
func hostAndPathFile(urlStr string) (string, error) {
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

// Deprecated: using urlutil
func host(urlStr string) (string, error) {

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	host := u.Host
	// if have port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return host, nil
}

// Deprecated: using urlutil
func GetPathFileNameFromUrl(prefixPath, url string) (pathFileName string, err error) {
	hostPathFile, err := hostAndPathFile(url)
	if err != nil {
		return "", err
	}
	return prefixPath + GetPathSeparator() + hostPathFile, nil
}

// Deprecated: using urlutil
func GetHostPathFromUrl(prefixPath, url string) (filePathName string, err error) {
	host, err := host(url)
	if err != nil {
		return "", err
	}
	return prefixPath + GetPathSeparator() + host, nil
}
