package urlutil

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	// 第三方依赖默认可用
)

// -------------------------- 私有辅助函数测试 --------------------------
func TestGetHostWithoutPort(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "带端口的IPv4",
			host:     "192.168.1.1:8080",
			expected: "192.168.1.1",
		},
		{
			name:     "带端口的域名",
			host:     "aa.com:80",
			expected: "aa.com",
		},
		{
			name:     "无端口的域名",
			host:     "aa.com",
			expected: "aa.com",
		},
		{
			name:     "IPv6地址（带端口）",
			host:     "[2001:db8::1]:8080",
			expected: "2001:db8::1",
		},
		{
			name:     "IPv6地址（无端口）",
			host:     "[2001:db8::1]",
			expected: "[2001:db8::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getHostWithoutPort(tt.host))
		})
	}
}

func TestProcessScheme(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		expected string
	}{
		{
			name:     "空scheme（兜底为http）",
			scheme:   "",
			expected: "http://",
		},
		{
			name:     "rsync scheme",
			scheme:   "rsync",
			expected: "rsync://",
		},
		{
			name:     "https scheme",
			scheme:   "https",
			expected: "https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &url.URL{Scheme: tt.scheme}
			assert.Equal(t, tt.expected, processScheme(u))
		})
	}
}

func TestCleanURLPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "含反斜杠的路径",
			path:     "\\repo\\test",
			expected: "/repo/test",
		},
		{
			name:     "连续斜杠的路径",
			path:     "//repo//test//",
			expected: "/repo/test",
		},
		{
			name:     "空路径",
			path:     "",
			expected: "/",
		},
		{
			name:     "根路径",
			path:     "/",
			expected: "/",
		},
		{
			name:     "不以/开头的路径",
			path:     "repo/test",
			expected: "/repo/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cleanURLPath(tt.path))
		})
	}
}

func TestJoinURLPath(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "拼接多个路径段",
			parts:    []string{"/root/path", "aa.com", "/repo/cc.html"},
			expected: "/root/path/aa.com/repo/cc.html",
		},
		{
			name:     "含空段的拼接",
			parts:    []string{"/root/", "", "aa.com//", "/repo"},
			expected: "/root/aa.com/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, joinURLPath(tt.parts...))
		})
	}
}

// -------------------------- 基础校验函数测试 --------------------------
func TestCheckUrlHost(t *testing.T) {
	tests := []struct {
		name    string
		urlStr  string
		wantErr bool
	}{
		{
			name:    "空URL",
			urlStr:  "",
			wantErr: true,
		},
		{
			name:    "仅空格的URL",
			urlStr:  "   ",
			wantErr: true,
		},
		{
			name:    "非法格式URL（空Host）",
			urlStr:  "http:///aa.com",
			wantErr: true,
		},
		{
			name:    "合法URL",
			urlStr:  "http://aa.com:8080/repo",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := checkUrlHost(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{
			name:    "空host",
			host:    "",
			wantErr: true,
		},
		{
			name:    "含非法字符的host（%）",
			host:    "aa%2Ecom",
			wantErr: true,
		},
		{
			name:    "非法域名（连续点）",
			host:    "aa..com",
			wantErr: true,
		},
		{
			name:    "合法IPv4（带端口）",
			host:    "192.168.1.1:8080",
			wantErr: false,
		},
		{
			name:    "合法IPv6（带端口）",
			host:    "[2001:db8::1]:8080",
			wantErr: false,
		},
		{
			name:    "合法域名",
			host:    "aa.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidHost(tt.host)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidURLPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "空path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "不以/开头的path",
			path:    "repo/test",
			wantErr: true,
		},
		{
			name:    "含路径遍历的path（../）",
			path:    "/../etc/passwd",
			wantErr: true,
		},
		{
			name:    "含非法字符的path（:）",
			path:    "/repo/test:file",
			wantErr: true,
		},
		{
			name:    "含非法转义的path（%2G）",
			path:    "/repo/%2Gtest",
			wantErr: true,
		},
		{
			name:    "合法path（多级）",
			path:    "/repo/defautl/xxxx",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidURLPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		urlStr  string
		wantErr bool
	}{
		{
			name:    "URL解析失败（空Host）",
			urlStr:  "http:///aa.com",
			wantErr: true,
		},
		{
			name:    "Host校验失败（非法字符）",
			urlStr:  "http://aa%2Ecom/repo",
			wantErr: true,
		},
		{
			name:    "合法URL",
			urlStr:  "http://aa.com:8080/repo",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAndValidateURL(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPreprocessAndValidatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "空path（预处理为/）",
			path:     "",
			expected: "/",
			wantErr:  false,
		},
		{
			name:     "非法path（不以/开头）",
			path:     "repo/test",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "合法path",
			path:     "/repo/test",
			expected: "/repo/test",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processedPath, err := preprocessAndValidatePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, processedPath)
			}
		})
	}
}

// -------------------------- 核心业务函数测试 --------------------------
func TestHostAndPort(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{
			name:     "带端口的URL",
			urlStr:   "http://aa.com:8080/repo",
			wantHost: "aa.com",
			wantPort: "8080",
			wantErr:  false,
		},
		{
			name:     "无端口的URL",
			urlStr:   "http://aa.com/repo",
			wantHost: "aa.com",
			wantPort: "",
			wantErr:  false,
		},
		{
			name:     "非法Host的URL",
			urlStr:   "http://aa%2Ecom/repo",
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := HostAndPort(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantHost, host)
				assert.Equal(t, tt.wantPort, port)
			}
		})
	}
}

func TestHostAndPath(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "合法URL（多级路径）",
			urlStr:   "http://aa.com:8080/repo/defautl/xxxx",
			expected: "aa.com/repo/defautl/",
			wantErr:  false,
		},
		{
			name:     "非法Path的URL（路径遍历）",
			urlStr:   "http://aa.com/repo/../test",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := HostAndPath(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestHost(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "带端口的URL",
			urlStr:   "http://aa.com:8080/repo",
			expected: "aa.com",
			wantErr:  false,
		},
		{
			name:     "非法Host的URL",
			urlStr:   "http://aa%2Ecom/repo",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Host(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "有Path的URL",
			urlStr:   "http://aa.com/repo/test",
			expected: "/repo/test",
			wantErr:  false,
		},
		{
			name:     "空Path的URL（预处理为/）",
			urlStr:   "http://aa.com",
			expected: "/",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Path(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestHostAndPathAndFile(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		wantHost string
		wantPath string
		wantFile string
		wantErr  bool
	}{
		{
			name:     "合法URL（带文件名）",
			urlStr:   "http://aa.com:8080/repo/bb/cc.html",
			wantHost: "aa.com",
			wantPath: "/repo/bb/",
			wantFile: "cc.html",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, path, file, err := HostAndPathAndFile(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantHost, host)
				assert.Equal(t, tt.wantPath, path)
				assert.Equal(t, tt.wantFile, file)
			}
		})
	}
}

func TestIsUrl(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected bool
	}{
		{
			name:     "合法http URL",
			urlStr:   "http://aa.com/repo",
			expected: true,
		},
		{
			name:     "合法file URL",
			urlStr:   "file:///root/test.txt",
			expected: true,
		},
		{
			name:     "空scheme URL",
			urlStr:   "//aa.com/repo",
			expected: false,
		},
		{
			name:     "非法Host的URL",
			urlStr:   "http://aa%2Ecom/repo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsUrl(tt.urlStr))
		})
	}
}

func TestSchemeAndHostAndPath(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "rsync URL（带端口）",
			urlStr:   "rsync://aa.com:8080/repo/defautl/xxxx",
			expected: "rsync://aa.com/repo/defautl/",
			wantErr:  false,
		},
		{
			name:     "空scheme URL（兜底为http）",
			urlStr:   "//aa.com/repo/test",
			expected: "http://aa.com/repo/",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := SchemeAndHostAndPath(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestSchemeAndHostAndFirstPath(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "rsync URL（多级路径）",
			urlStr:   "rsync://aa.com:8080/repo/defautl/xxxx",
			expected: "rsync://aa.com/repo/",
			wantErr:  false,
		},
		{
			name:     "http URL（空Path）",
			urlStr:   "http://aa.com",
			expected: "http://aa.com/",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := SchemeAndHostAndFirstPath(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestHostAndPathFile(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "合法URL（带文件名）",
			urlStr:   "http://aa.com:8080/repo/cc.html",
			expected: "aa.com/repo/cc.html", // 核心：无开头/
			wantErr:  false,
		},
		{
			name:     "非法Path的URL",
			urlStr:   "http://aa.com/repo/test:file",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := hostAndPathFile(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestJoinPrefixPathAndUrlFileName(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "合法拼接",
			prefix:   "/root/path",
			urlStr:   "http://aa.com/repo/cc.html",
			expected: "/root/path/aa.com/repo/cc.html",
			wantErr:  false,
		},
		{
			name:     "空prefix",
			prefix:   "",
			urlStr:   "http://aa.com/repo/cc.html",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := JoinPrefixPathAndUrlFileName(tt.prefix, tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestJoinPrefixPathAndUrlHost(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		urlStr   string
		expected string
		wantErr  bool
	}{
		{
			name:     "合法拼接",
			prefix:   "/root/path",
			urlStr:   "http://aa.com:8080/repo",
			expected: "/root/path/aa.com",
			wantErr:  false,
		},
		{
			name:     "非法URL",
			prefix:   "/root/path",
			urlStr:   "http://aa%2Ecom/repo",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := JoinPrefixPathAndUrlHost(tt.prefix, tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestHasPort(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected bool
	}{
		{
			name:     "带有效端口",
			addr:     "aa.com:8080",
			expected: true,
		},
		{
			name:     "末尾冒号（无效端口）",
			addr:     "192.168.1.1:",
			expected: false,
		},
		{
			name:     "无端口",
			addr:     "aa.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasPort(tt.addr))
		})
	}
}

func TestTryJoinHostPort(t *testing.T) {
	tests := []struct {
		name     string
		server   string
		port     string
		expected string
	}{
		{
			name:     "已有端口",
			server:   "aa.com:8080",
			port:     "80",
			expected: "aa.com:8080",
		},
		{
			name:     "IPv4末尾冒号",
			server:   "192.168.1.1:",
			port:     "8080",
			expected: "192.168.1.1:8080",
		},
		{
			name:     "无端口拼接",
			server:   "aa.com",
			port:     "8080",
			expected: "aa.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, TryJoinHostPort(tt.server, tt.port))
		})
	}
}
