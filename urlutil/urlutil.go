package urlutil

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/cpusoft/goutil/osutil"
)

// http://server:port/aa/bb/cc.html --> server port
// http://server/aa/bb/cc.html --> server ""
func HostAndPort(urlStr string) (host, port string, err error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", "", err
	}

	host, port, err = net.SplitHostPort(u.Host)
	if err != nil {
		return u.Host, "", nil
	}
	return host, port, nil
}

// http://server:port/aa/bb/cc.html --> server/aa/bb/
func HostAndPath(urlStr string) (string, error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", err
	}

	processedPath, err := preprocessAndValidatePath(u.Path)
	if err != nil {
		return "", err
	}

	host := getHostWithoutPort(u.Host)
	decodedHost, err := url.PathUnescape(host)
	if err != nil {
		return "", err
	}

	decodedPath, err := url.PathUnescape(processedPath)
	if err != nil {
		return "", err
	}

	pos := strings.LastIndex(decodedPath, "/")
	if pos == -1 {
		return decodedHost + "/" + decodedPath + "/", nil
	}
	return decodedHost + decodedPath[:pos+1], nil
}

// http://server:port/aa/bb/cc.html --> server
func Host(urlStr string) (string, error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", err
	}

	return getHostWithoutPort(u.Host), nil
}

// http://server:port/aa/bb/cc.html --> /aa/bb/cc.html
func Path(urlStr string) (string, error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", err
	}

	processedPath, err := preprocessAndValidatePath(u.Path)
	if err != nil {
		return "", err
	}

	return processedPath, nil
}

// http://server:port/aa/bb/cc.html --> server,  /aa/bb, cc.html
func HostAndPathAndFile(urlStr string) (host, path, file string, err error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", "", "", err
	}

	processedPath, err := preprocessAndValidatePath(u.Path)
	if err != nil {
		return "", "", "", err
	}

	host = getHostWithoutPort(u.Host)
	path, file = osutil.Split(processedPath)
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

	if u.Scheme == "" {
		return false
	}

	if u.Scheme != "file" {
		if u.Host == "" || isValidHost(u.Host) != nil {
			return false
		}
		return true
	}

	cleanPath := strings.TrimSpace(u.Path)
	if cleanPath == "" || cleanPath == "/" || isValidURLPath(cleanPath) != nil {
		return false
	}
	return true
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
	if port == "" {
		return false
	}
	return true
}

// http://server:port/aa/bb/cc.html --> http://server/aa/bb/
func SchemeAndHostAndPath(urlStr string) (string, error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", err
	}

	processedPath, err := preprocessAndValidatePath(u.Path)
	if err != nil {
		return "", err
	}

	scheme := processScheme(u)
	host := getHostWithoutPort(u.Host)

	pos := strings.LastIndex(processedPath, "/")
	if pos == -1 {
		return scheme + host + "/" + processedPath + "/", nil
	}
	return scheme + host + processedPath[:pos+1], nil
}

// rsync://aa.com:xxx/repo/defautl/xxxx --> rsync://aa.com/repo/
func SchemeAndHostAndFirstPath(urlStr string) (string, error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", err
	}

	processedPath, err := preprocessAndValidatePath(u.Path)
	if err != nil {
		return "", err
	}

	scheme := processScheme(u)
	host := getHostWithoutPort(u.Host)

	// 修正：改用URL专用cleanURLPath，避免系统分隔符问题
	cleanPath := cleanURLPath(processedPath)
	// 按/拆分路径（纯字符串处理，无系统差异）
	split := strings.Split(cleanPath, "/")
	// 过滤空字符串（如/拆分后是["", ""]）
	var validParts []string
	for _, part := range split {
		if part != "" {
			validParts = append(validParts, part)
		}
	}

	var firstPath string
	if len(validParts) == 0 {
		firstPath = "/" // 空Path返回/
	} else {
		firstPath = "/" + validParts[0] + "/" // 取第一个有效路径段
	}

	return scheme + host + firstPath, nil
}

// url is http://www.server.com:8080/aa/bb/cc.html , and  prefixPath is /root/path ;
// combine to  /root/path/www.server.com/aa/bb/cc.html
func JoinPrefixPathAndUrlFileName(prefixPath, urlStr string) (filePathName string, err error) {
	trimmedPrefix := strings.TrimSpace(prefixPath)
	if trimmedPrefix == "" {
		return "", errors.New("prefixPath is empty or only contains whitespace")
	}

	hostPathFile, err := hostAndPathFile(urlStr)
	if err != nil {
		return "", err
	}

	decodedHostPathFile, err := url.PathUnescape(hostPathFile)
	if err != nil {
		return "", err
	}
	// 修正1：移除filepath.Clean，改用cleanURLPath
	escapedHostPathFile := cleanURLPath(decodedHostPathFile)
	// 修正2：替换osutil.JoinPathFile/filepath.Clean为joinURLPath
	filePathName = joinURLPath(trimmedPrefix, escapedHostPathFile)
	return filePathName, nil
}

// url is http://www.server.com:8080/aa/bb/cc.html , and  preifxPath is /root/path ;
// combine to  /root/path/www.server.com
func JoinPrefixPathAndUrlHost(prefixPath, urlStr string) (path string, err error) {
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
	// 修正1：移除filepath.Clean，改用cleanURLPath
	escapedHost := cleanURLPath(decodedHostPathFile)
	// 修正2：替换osutil.JoinPathFile/filepath.Clean为joinURLPath
	path = joinURLPath(trimmedPrefix, escapedHost)
	return path, nil
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
	if strings.HasSuffix(server, ":") {
		trimmed := server[:len(server)-1]
		if ip := net.ParseIP(trimmed); ip != nil && ip.To4() != nil {
			sanitized = trimmed
		}
	}

	return net.JoinHostPort(sanitized, port)
}

// http://server:port/aa/bb/cc.html --> server/aa/bb/cc.html
func hostAndPathFile(urlStr string) (string, error) {
	u, err := parseAndValidateURL(urlStr)
	if err != nil {
		return "", err
	}

	processedPath, err := preprocessAndValidatePath(u.Path)
	if err != nil {
		return "", err
	}

	host := getHostWithoutPort(u.Host)
	decodedHost, err := url.PathUnescape(host)
	if err != nil {
		return "", err
	}
	// 核心修正：手动拼接host和path，而非用joinURLPath（避免host前加/）
	cleanPath := cleanURLPath(processedPath)
	if strings.HasPrefix(cleanPath, "/") {
		return decodedHost + cleanPath, nil
	}
	return decodedHost + "/" + cleanPath, nil
}

// filepath.clean在windows下有问题，改为cleanURLPath 清理URL路径，统一分隔符为/，移除连续/，符合URL规范
func cleanURLPath(path string) string {
	// 1. 统一反斜杠为正斜杠（消除系统差异）
	path = strings.ReplaceAll(path, "\\", "/")
	// 2. 清理连续的/（避免//、///等异常）
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	// 3. 空路径返回/（URL根路径规范）
	if path == "" {
		return "/"
	}
	// 4. 确保路径以/开头（URL路径规范）
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// 5. 移除末尾的/（除非是根路径）
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

// joinURLPath 拼接URL路径段，统一用/分隔，避免系统差异
func joinURLPath(parts ...string) string {
	var result string
	for _, part := range parts {
		if part == "" {
			continue
		}
		// 清理当前段的分隔符
		part = cleanURLPath(part)
		// 拼接逻辑：确保段之间只有一个/
		if result == "" {
			result = part
		} else {
			if strings.HasSuffix(result, "/") {
				if strings.HasPrefix(part, "/") {
					result += part[1:]
				} else {
					result += part
				}
			} else {
				if strings.HasPrefix(part, "/") {
					result += part
				} else {
					result += "/" + part
				}
			}
		}
	}
	// 最终清理一次，确保规范
	return cleanURLPath(result)
}

func getHostWithoutPort(host string) string {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		// 无端口时返回原host（如IPv6地址[2001:db8::1]）
		return host
	}
	return h
}

func parseAndValidateURL(urlStr string) (*url.URL, error) {
	u, err := checkUrlHost(urlStr)
	if err != nil {
		return nil, fmt.Errorf("URL基础解析失败: %w", err)
	}

	if err := isValidHost(u.Host); err != nil {
		return nil, fmt.Errorf("Host校验失败: %w", err)
	}

	return u, nil
}

func preprocessAndValidatePath(path string) (string, error) {
	processedPath := path
	if processedPath == "" {
		processedPath = "/"
	}

	if err := isValidURLPath(processedPath); err != nil {
		return "", fmt.Errorf("Path校验失败: %w", err)
	}

	return processedPath, nil
}

func processScheme(u *url.URL) string {
	scheme := u.Scheme
	if scheme == "" {
		scheme = "http" // 兜底：默认http
	}
	return scheme + "://"
}

func checkUrlHost(urlStr string) (*url.URL, error) {
	trimmedUrl := strings.TrimSpace(urlStr)
	if trimmedUrl == "" {
		return nil, errors.New("URL is empty or only contains whitespace")
	}

	u, err := url.Parse(trimmedUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %w", trimmedUrl, err)
	}
	if len(u.Host) == 0 {
		return nil, fmt.Errorf("URL '%s' host is empty", trimmedUrl)
	}
	return u, nil
}

func isValidHost(host string) error {
	if strings.TrimSpace(host) == "" {
		return errors.New("host is empty")
	}

	invalidChars := regexp.MustCompile(`[^\w\.\-:\[\]]`)
	if invalidChars.MatchString(host) {
		return fmt.Errorf("host '%s' contains illegal characters (only letters, digits, -, ., :, [, ] are allowed)", host)
	}

	hostOnly, _, err := net.SplitHostPort(host)
	if err != nil {
		hostOnly = host
	}

	if ip := net.ParseIP(hostOnly); ip != nil {
		return nil
	}

	validDomainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	if !validDomainRegex.MatchString(hostOnly) {
		return fmt.Errorf("host '%s' is not a valid domain (RFC 1035)", host)
	}

	return nil
}

func isValidURLPath(path string) error {
	if path == "" {
		return errors.New("URL path is empty")
	}

	if !utf8.ValidString(path) {
		return errors.New("URL path contains invalid UTF-8 characters")
	}

	if _, err := url.PathUnescape(path); err != nil {
		return fmt.Errorf("URL path '%s' contains invalid escape sequences: %w", path, err)
	}

	decodedPath, _ := url.PathUnescape(path)
	if strings.Contains(decodedPath, "../") || strings.Contains(decodedPath, "./") {
		return fmt.Errorf("URL path '%s' contains path traversal sequences (../ or ./)", path)
	}

	invalidPathChars := regexp.MustCompile(`[\\:*?"<>|]`)
	if invalidPathChars.MatchString(decodedPath) {
		return fmt.Errorf("URL path '%s' contains illegal filename characters (\\ : * ? \" < > |)", path)
	}

	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("URL path '%s' must start with /", path)
	}

	return nil
}
