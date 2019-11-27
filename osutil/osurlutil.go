package osutil

import (
	urlutil "github.com/cpusoft/goutil/urlutil"
)

// url is http://www.server.com:8080/aa/bb/cc.html , and  preifxPath is /root/path ;
// combine to  /root/path/www.server.com/aa/bb/cc.html
func GetPathFileNameFromUrl(prefixPath, url string) (pathFileName string, err error) {
	hostPathFile, err := urlutil.HostAndPathFile(url)
	if err != nil {
		return "", err
	}
	return prefixPath + GetPathSeparator() + hostPathFile, nil
}

// url is http://www.server.com:8080/aa/bb/cc.html , and  preifxPath is /root/path ;
// combine to  /root/path/www.server.com
func GetHostPathFromUrl(prefixPath, url string) (filePathName string, err error) {
	host, err := urlutil.Host(url)
	if err != nil {
		return "", err
	}
	return prefixPath + GetPathSeparator() + host, nil
}
