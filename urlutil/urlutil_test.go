package urlutil

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试公共函数：GetHostWithoutPort（IPv4/IPv6/带端口/无端口）
func TestGetHostWithoutPort(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"IPv4带端口", "192.168.1.1:8080", "192.168.1.1"},
		{"IPv4无端口", "192.168.1.1", "192.168.1.1"},
		{"IPv6带端口", "[2001:db8::1]:8080", "2001:db8::1"},
		{"IPv6无端口", "[2001:db8::1]", "[2001:db8::1]"},
		{"域名带端口", "example.com:80", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetHostWithoutPort(tt.host))
		})
	}
}

// 测试checkUrlHost（空URL/空格URL/解析失败/host为空）
func TestCheckUrlHost(t *testing.T) {
	tests := []struct {
		name    string
		urlStr  string
		wantErr bool
		errMsg  string
	}{
		{"空URL", "", true, "URL is empty or only contains whitespace"},
		{"空格URL", "   ", true, "URL is empty or only contains whitespace"},
		{"解析失败URL", "abc123://", true, "parse URL 'abc123://' failed:"},
		{"host为空URL", "http://", true, "URL 'http://' host is empty"},
		{"合法URL", "http://example.com:8080/test", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := checkUrlHost(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 测试HostAndPort（IPv4/IPv6/带端口/无端口/编码host）
func TestHostAndPort(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{"IPv4带端口", "http://192.168.1.1:8080/test", "192.168.1.1", "8080", false},
		{"IPv4无端口", "http://192.168.1.1/test", "192.168.1.1", "", false},
		{"IPv6带端口", "http://[2001:db8::1]:8080/test", "2001:db8::1", "8080", false},
		{"IPv6无端口", "http://[2001:db8::1]/test", "[2001:db8::1]", "", false},
		{"编码host", "http://example%2Ecom:80/test", "example%2Ecom", "80", false},
		{"非法URL", "abc123", "", "", true},
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

// 测试HostAndPath（编码host/编码path/空path/无/的path）
func TestHostAndPath(t *testing.T) {
	tests := []struct {
		name    string
		urlStr  string
		want    string
		wantErr bool
	}{
		{"编码host+编码path", "http://example%2Ecom:80/%20test.txt", "example.com/ /", false},
		{"空path", "http://example.com", "example.com/", false},
		{"无/的path", "http://example.com/testfile", "example.com/testfile/", false},
		{"多级path", "http://example.com/aa/bb/cc.html", "example.com/aa/bb/", false},
		{"非法URL", "abc123", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HostAndPath(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// 测试SchemeAndHostAndPath（编码host/编码path/空scheme）
func TestSchemeAndHostAndPath(t *testing.T) {
	tests := []struct {
		name    string
		urlStr  string
		want    string
		wantErr bool
	}{
		{"空scheme", "//example.com/aa/bb.html", "http://example.com/aa/", false},
		{"编码host+path", "https://example%2Ecom/%20test.txt", "https://example%2Ecom/ /", false},
		{"IPv6+空path", "http://[2001:db8::1]", "http://[2001:db8::1]/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SchemeAndHostAndPath(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// 测试SchemeAndHostAndFirstPath（Windows路径/编码path/空path）
func TestSchemeAndHostAndFirstPath(t *testing.T) {
	// 模拟Windows路径（替换分隔符）
	windowsPathUrl := "rsync://example.com:\\repo\\defautl\\xxxx"
	wantWindows := "rsync://example.com/repo/"

	tests := []struct {
		name    string
		urlStr  string
		want    string
		wantErr bool
	}{
		{"Windows路径", windowsPathUrl, wantWindows, false},
		{"编码path", "rsync://example.com/%20repo/defautl", "rsync://example.com/ repo/", false},
		{"空path", "rsync://example.com", "rsync://example.com//", false}, // 空path返回//，符合逻辑
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SchemeAndHostAndFirstPath(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// 测试HostAndPathFile（编码host/拼接分隔符/IPv6）
func TestHostAndPathFile(t *testing.T) {
	tests := []struct {
		name    string
		urlStr  string
		want    string
		wantErr bool
	}{
		{"编码host", "http://example%2Ecom:80/aa/bb.html", "example.com/aa/bb.html", false},
		{"host带/分隔符", "http://example.com/:80/aa.html", "example.com:/aa.html", false}, // filepath.Join自动处理/
		{"IPv6", "http://[2001:db8::1]/aa/bb.html", "[2001:db8::1]/aa/bb.html", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HostAndPathFile(tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// 测试HostAndPathAndFile（编码host/编码path/空path）
func TestHostAndPathAndFile(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		wantHost string
		wantPath string
		wantFile string
		wantErr  bool
	}{
		{"编码host+path", "http://example%2Ecom/%20test.txt", "example.com", "/ ", "test.txt", false},
		{"空path", "http://example.com", "example.com", "/", "", false},
		{"多级path", "http://example.com/aa/bb/cc.html", "example.com", "/aa/bb/", "cc.html", false},
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

// 测试IsUrl（file协议/空scheme/空host/合法URL）
func TestIsUrl(t *testing.T) {
	tests := []struct {
		name   string
		urlStr string
		want   bool
	}{
		{"合法http URL", "http://example.com/test", true},
		{"合法file URL", "file:///etc/hosts", true},
		{"无效file URL", "file://", false},
		{"空scheme", "//example.com", false},
		{"空host", "http://", false},
		{"空格URL", "   ", false},
		{"相对URL", "/aa/bb.html", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsUrl(tt.urlStr))
		})
	}
}

// 测试JoinPrefixPathAndUrlFileName（空prefix/空格prefix/编码路径）
func TestJoinPrefixPathAndUrlFileName(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name       string
		prefixPath string
		urlStr     string
		want       string
		wantErr    bool
		errMsg     string
	}{
		{"空prefix", "", "http://example.com/test.txt", "", true, "prefixPath is empty or only contains whitespace"},
		{"空格prefix", "   ", "http://example.com/test.txt", "", true, "prefixPath is empty or only contains whitespace"},
		{"编码路径", tmpDir, "http://example%2Ecom/%20test.txt", filepath.Join(tmpDir, "example.com/ test.txt"), false, ""},
		{"合法拼接", tmpDir, "http://example.com/aa/bb.html", filepath.Join(tmpDir, "example.com/aa/bb.html"), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JoinPrefixPathAndUrlFileName(tt.prefixPath, tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// 测试JoinPrefixPathAndUrlHost（空prefix/空格prefix/编码host）
func TestJoinPrefixPathAndUrlHost(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name       string
		prefixPath string
		urlStr     string
		want       string
		wantErr    bool
		errMsg     string
	}{
		{"空prefix", "", "http://example.com/test.txt", "", true, "prefixPath is empty or only contains whitespace"},
		{"空格prefix", "   ", "http://example.com/test.txt", "", true, "prefixPath is empty or only contains whitespace"},
		{"编码host", tmpDir, "http://example%2Ecom:80/test", filepath.Join(tmpDir, "example.com"), false, ""},
		{"IPv6 host", tmpDir, "http://[2001:db8::1]/test", filepath.Join(tmpDir, "[2001:db8::1]"), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JoinPrefixPathAndUrlHost(tt.prefixPath, tt.urlStr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// 测试HasPort（IPv4带冒号/IPv6带端口/空port）
func TestHasPort(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{"IPv4带端口", "192.168.1.1:8080", true},
		{"IPv4末尾冒号", "192.168.1.1:", false}, // 空port返回false
		{"IPv6带端口", "[2001:db8::1]:8080", true},
		{"IPv6无端口", "[2001:db8::1]", false},
		{"域名无端口", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasPort(tt.addr))
		})
	}
}

// 测试TryJoinHostPort（空server/空port/IPv4末尾冒号/IPv6错误拼接）
func TestTryJoinHostPort(t *testing.T) {
	tests := []struct {
		name   string
		server string
		port   string
		want   string
	}{
		{"空server", "", "8080", ""},
		{"空port", "192.168.1.1", "", "192.168.1.1"},
		{"IPv4末尾冒号", "192.168.1.1:", "8080", "192.168.1.1:8080"},
		{"IPv6末尾冒号", "[::1]:", "8080", "[::1]::8080"}, // 漏洞场景：生成无效地址
		{"IPv6合法拼接", "[::1]", "8080", "[::1]:8080"},
		{"已带端口", "example.com:80", "8080", "example.com:80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TryJoinHostPort(tt.server, tt.port))
		})
	}
}

func TestHost1(t *testing.T) {
	url := `//1.2.3.4:33`
	u, err := Host(url)
	fmt.Println(u, err)
}

func TestHost(t *testing.T) {
	url := `rsync://rpki.apnic.net:999/member_repository/A91270E6/75648ECED63511E896631322C4F9AE02/dVNRzYJvKfhxtLyVlPTpSNvnc-k.mft?aa=bbb`
	u, err := Host(url)
	fmt.Println(u, err)
}

func TestPath(t *testing.T) {
	url := `rsync://rpki.apnic.net:999/member_repository/A91270E6/75648ECED63511E896631322C4F9AE02/dVNRzYJvKfhxtLyVlPTpSNvnc-k.mft?aa=bbb`
	u, err := Path(url)
	fmt.Println(u, err)
}
