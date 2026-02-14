package urlutil

import (
	"errors"
	"net"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cpusoft/goutil/osutil"
)

func GetHostWithoutPort(host string) string {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		// 无端口时返回原host（如IPv6地址[2001:db8::1]）
		return host
	}
	return h
}

func checkUrlHost(urlStr string) (*url.URL, error) {
	// 修复1：处理空格字符串
	trimmedUrl := strings.TrimSpace(urlStr)
	if trimmedUrl == "" {
		return nil, errors.New("URL is empty or only contains whitespace")
	}

	u, err := url.Parse(trimmedUrl)
	if err != nil {
		return nil, err
	}
	if len(u.Host) == 0 {
		return nil, errors.New("URL host is empty")
	}
	return u, nil
}

// http://server:port/aa/bb/cc.html --> server port
// http://server/aa/bb/cc.html --> server ""
func HostAndPort(urlStr string) (host, port string, err error) {
	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", "", err
	}
	host, port, err = net.SplitHostPort(u.Host)
	if err != nil {
		// 无端口时返回原host和空port
		return u.Host, "", nil
	}
	return host, port, nil
}

// http://server:port/aa/bb/cc.html --> server/aa/bb/
func HostAndPath(urlStr string) (string, error) {

	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", err
	}
	host := GetHostWithoutPort(u.Host)
	decodedHost, err := url.PathUnescape(host)
	if err != nil {
		return "", err
	}

	// 处理空Path/无/的Path
	path := u.Path
	// 解码path中的URL编码字符
	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return "", err
	}
	if decodedPath == "" {
		return decodedHost + "/", nil
	}
	pos := strings.LastIndex(decodedPath, "/")
	if pos == -1 {
		return decodedHost + "/" + decodedPath + "/", nil
	}
	return decodedHost + decodedPath[:pos+1], nil
}

// http://server:port/aa/bb/cc.html --> http://server/aa/bb/
func SchemeAndHostAndPath(urlStr string) (string, error) {
	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", err
	}

	scheme := u.Scheme
	if scheme == "" {
		scheme = "http" // 兜底：默认http
	}
	scheme += "://"

	// 用公共函数获取无端口的host
	host := GetHostWithoutPort(u.Host)

	// 处理空Path/无/的Path
	path := u.Path
	if path == "" {
		return scheme + host + "/", nil
	}
	pos := strings.LastIndex(path, "/")
	if pos == -1 {
		return scheme + host + "/" + path + "/", nil
	}
	return scheme + host + path[:pos+1], nil
}

// rsync://aa.com:xxx/repo/defautl/xxxx --> rsync://aa.com/repo/
func SchemeAndHostAndFirstPath(urlStr string) (string, error) {
	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", err
	}
	scheme := u.Scheme
	if scheme == "" {
		scheme = "http" // 兜底：默认http
	}
	scheme += "://"

	// 用公共函数获取无端口的host
	host := GetHostWithoutPort(u.Host)

	// 路径拆分逻辑（处理开头/、空Path、多级空路径）
	path := u.Path
	// 清理冗余分隔符，统一路径格式
	cleanPath := filepath.Clean(strings.ReplaceAll(path, "\\", "/"))
	if cleanPath == "." {
		// 空Path，返回/
		cleanPath = "/"
	}
	split := strings.FieldsFunc(cleanPath, func(r rune) bool { return r == '/' })

	var firstPath string
	if len(split) == 0 {
		firstPath = "/"
	} else {
		// 拼接为首路径段（如repo → /repo/）
		firstPath = "/" + split[0] + "/"
	}

	return scheme + host + firstPath, nil
}

// http://server:port/aa/bb/cc.html --> server
func Host(urlStr string) (string, error) {

	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", err
	}
	return GetHostWithoutPort(u.Host), nil
}

// http://server:port/aa/bb/cc.html --> /aa/bb/cc.html
func Path(urlStr string) (string, error) {
	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", err
	}
	path := u.Path
	if path == "" {
		return "/", nil // 兜底：空Path返回/
	}
	return path, nil
}

// http://server:port/aa/bb/cc.html --> server/aa/bb/cc.html
func HostAndPathFile(urlStr string) (string, error) {
	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", err
	}
	host := GetHostWithoutPort(u.Host)
	// 修复1：解码host中的URL编码字符
	decodedHost, err := url.PathUnescape(host)
	if err != nil {
		return "", err
	}
	// 修复2：用filepath.Join拼接，自动处理分隔符
	return filepath.Join(decodedHost, u.Path), nil
}

// http://server:port/aa/bb/cc.html --> server,  aa/bb, cc.html
func HostAndPathAndFile(urlStr string) (host, path, file string, err error) {
	u, err := checkUrlHost(urlStr)
	if err != nil {
		return "", "", "", err
	}
	// 用公共函数获取无端口的host
	host = GetHostWithoutPort(u.Host)

	// 空Path处理
	pathStr := u.Path
	if pathStr == "" {
		pathStr = "/"
	}
	path, file = osutil.Split(pathStr)
	return host, path, file, nil
}

func IsUrl(urlStr string) bool {
	trimmedUrl := strings.TrimSpace(urlStr)
	if trimmedUrl == "" {
		return false
	}
	u, err := url.Parse(trimmedUrl)
	if err != nil {
		return false
	}
	// 合法URL规则：
	// 1. scheme非空；
	// 2. 非file协议：host非空；
	// 3. file协议：path非空（且不能仅为/）
	if u.Scheme == "" {
		return false
	}
	if u.Scheme != "file" {
		return u.Host != ""
	}
	// file协议需满足path非空且不是仅为/
	cleanPath := strings.TrimSpace(u.Path)
	return cleanPath != "" && cleanPath != "/"
}

// url is http://www.server.com:8080/aa/bb/cc.html , and  prefixPath is /root/path ;
// combine to  /root/path/www.server.com/aa/bb/cc.html
func JoinPrefixPathAndUrlFileName(prefixPath, urlStr string) (filePathName string, err error) {
	trimmedPrefix := strings.TrimSpace(prefixPath)
	if trimmedPrefix == "" {
		return "", errors.New("prefixPath is empty or only contains whitespace")
	}

	hostPathFile, err := HostAndPathFile(urlStr)
	if err != nil {
		return "", err
	}
	// 修复：先解码URL编码字符，再清理路径
	decodedHostPathFile, err := url.PathUnescape(hostPathFile)
	if err != nil {
		return "", err
	}
	// 转义特殊字符，防止路径遍历
	escapedHostPathFile := filepath.Clean(decodedHostPathFile)
	filePathName = osutil.JoinPathFile(trimmedPrefix, escapedHostPathFile)
	// 清理冗余分隔符
	filePathName = filepath.Clean(filePathName)
	return filePathName, nil
}

// url is http://www.server.com:8080/aa/bb/cc.html , and  prefixPath is /root/path ;
// combine to  /root/path/www.server.com
func JoinPrefixPathAndUrlHost(prefixPath, urlStr string) (path string, err error) {
	// 新增：空值/空格校验
	trimmedPrefix := strings.TrimSpace(prefixPath)
	if trimmedPrefix == "" {
		return "", errors.New("prefixPath is empty or only contains whitespace")
	}
	host, err := Host(urlStr)
	if err != nil {
		return "", err
	}
	decodedHostPathFile, err := url.PathUnescape(host)
	if err != nil {
		return "", err
	}
	// 转义特殊字符
	escapedHost := filepath.Clean(decodedHostPathFile)
	path = osutil.JoinPathFile(trimmedPrefix, escapedHost)
	path = filepath.Clean(path)
	return path, nil
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
	if server == "" || port == "" {
		return server
	}
	if HasPort(server) {
		return server
	}

	sanitized := server
	// 修正IPv4冒号截断逻辑（仅对IPv4且末尾单冒号的场景截断）
	if strings.HasSuffix(server, ":") {
		// 提取截断后的地址，判断是否为合法IPv4
		trimmed := server[:len(server)-1]
		if ip := net.ParseIP(trimmed); ip != nil && ip.To4() != nil {
			sanitized = trimmed
		}
	}

	return net.JoinHostPort(sanitized, port)
}
