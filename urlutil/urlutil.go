package urlutil

import (
	"net/url"
	"strings"

	"github.com/cpusoft/goutil/osutil"
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

// http://server:port/aa/bb/cc.html --> server
func Host(urlStr string) (string, error) {

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

// http://server:port/aa/bb/cc.html --> /aa/bb/cc.html
func Path(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	path := u.Path
	return path, nil
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

// http://server:port/aa/bb/cc.html --> server,  aa/bb, cc.html
func HostAndPathAndFile(urlStr string) (host, path, file string, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	host = u.Host
	// if have port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	path, file = osutil.Split(u.Path)
	return host, path, file, nil
}

func IsUrl(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}

// url is http://www.server.com:8080/aa/bb/cc.html , and  preifxPath is /root/path ;
// combine to  /root/path/www.server.com/aa/bb/cc.html
func JoinPrefixPathAndUrlFileName(prefixPath, url string) (pathFileName string, err error) {
	hostPathFile, err := HostAndPathFile(url)
	if err != nil {
		return "", err
	}
	return osutil.JoinPathFile(prefixPath, hostPathFile), nil
}

// url is http://www.server.com:8080/aa/bb/cc.html , and  preifxPath is /root/path ;
// combine to  /root/path/www.server.com
func JoinPrefixPathAndUrlHost(prefixPath, url string) (path string, err error) {
	host, err := Host(url)
	if err != nil {
		return "", err
	}
	return osutil.JoinPathFile(prefixPath, host), nil
}
