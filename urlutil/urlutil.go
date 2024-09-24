package urlutil

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/cpusoft/goutil/osutil"
)

// http://server:port/aa/bb/cc.html --> server port
// http://server/aa/bb/cc.html --> server ""
func HostAndPort(urlStr string) (host, port string, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}
	if len(u.Host) == 0 {
		return "", "", errors.New("it is not in a legal URL format")
	}
	if strings.Contains(u.Host, ":") {
		return net.SplitHostPort(u.Host)
	}
	return u.Host, "", nil
}

// http://server:port/aa/bb/cc.html --> server/aa/bb/
func HostAndPath(urlStr string) (string, error) {

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	if len(u.Host) == 0 {
		return "", errors.New("it is not in a legal URL format")
	}
	pos := strings.LastIndex(u.Path, "/")
	host := u.Host
	// if have port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return (host + string(u.Path[:pos+1])), nil
}

// http://server:port/aa/bb/cc.html --> http://server/aa/bb/
func SchemeAndHostAndPath(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	if len(u.Host) == 0 {
		return "", errors.New("it is not in a legal URL format")
	}
	scheme := u.Scheme + "://"
	pos := strings.LastIndex(u.Path, "/")
	host := u.Host
	// if have port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return (scheme + host + string(u.Path[:pos+1])), nil
}

// rsync://aa.com:xxx/repo/defautl/xxxx --> rsync://aa.com/repo/
func SchemeAndHostAndFirstPath(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	if len(u.Host) == 0 {
		return "", errors.New("it is not in a legal URL format")
	}
	scheme := u.Scheme + "://"
	host := u.Host
	// if have port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	path := u.Path
	split := strings.Split(path, `/`)
	//belogs.Debug("SchemeAndHostAndFirstPath(): urlStr:", urlStr, " path:", path, "  split:", jsonutil.MarshalJson(split))
	firstPath := split[0] + `/`
	if len(split) > 1 {
		firstPath = firstPath + split[1] + `/`
	}

	return (scheme + host + firstPath), nil
}

// http://server:port/aa/bb/cc.html --> server
func Host(urlStr string) (string, error) {

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	if len(u.Host) == 0 {
		return "", errors.New("it is not in a legal URL format")
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
	if len(u.Host) == 0 {
		return "", errors.New("it is not in a legal URL format")
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
	if len(u.Host) == 0 {
		return "", errors.New("it is not in a legal URL format")
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
	if len(u.Host) == 0 {
		return "", "", "", errors.New("it is not in a legal URL format")
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
func JoinPrefixPathAndUrlFileName(prefixPath, url string) (filePathName string, err error) {
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

// HostPort returns whether addr includes a port number (i.e.,
// is of the form HOST:PORT).  It handles a corner-case in [net.SplitHostPort]
// which returns an empty port for addresses of the form "1.2.3.4:".  For such
// addresses, HasPort returns false.
func HasPort(addr string) bool {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}

	// this deals with the corner-case of, e.g., "1.2.3.4:".  For
	// addresses of this form, net.SplitHostAddr does not return an
	// error, and returns an empty port string.
	if port == "" {
		return false
	}
	return true
}

// TryJoinHostPort checks whether the server string already has a port (i.e.,
// ends with ':<PORT>'.  If it does, then the function simply returns
// that string.  If it does not, it returns the server string with
// the port appended.
func TryJoinHostPort(server string, port string) string {
	if HasPort(server) {
		return server
	}

	sanitized := server
	if strings.HasSuffix(server, ":") && !strings.HasSuffix(server, "::") {
		sanitized = server[:len(server)-1]
	}

	return net.JoinHostPort(sanitized, port)
}
