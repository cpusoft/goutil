package whoisutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
	"github.com/stretchr/testify/assert"
)

// ===================== 基础工具函数测试 =====================
// TestWhoisConfig_getParamsWithQuery 测试参数拼接逻辑（核心修复点）
func TestWhoisConfig_getParamsWithQuery(t *testing.T) {
	tests := []struct {
		name    string
		config  *WhoisConfig
		query   string
		want    []string
		wantErr bool
	}{
		{
			name:   "正常配置（Host+Port）",
			config: &WhoisConfig{Host: "whois.cymru.com", Port: "43"},
			query:  "-v AS23028",
			want:   []string{"-h", "whois.cymru.com", "-p", "43", "-v AS23028"},
		},
		{
			name:   "仅Host",
			config: &WhoisConfig{Host: "whois.apnic.net"},
			query:  "8.8.8.8",
			want:   []string{"-h", "whois.apnic.net", "8.8.8.8"},
		},
		{
			name:   "仅Port",
			config: &WhoisConfig{Port: "8080"},
			query:  "2001:db8::1",
			want:   []string{"-p", "8080", "2001:db8::1"},
		},
		{
			name:   "空配置",
			config: &WhoisConfig{},
			query:  "AS12345",
			want:   []string{"AS12345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.getParamsWithQuery(tt.query)
			assert.Equal(t, tt.want, got, "参数拼接结果不符")
		})
	}
}

// TestNewWhoisResult 测试单行whois解析逻辑（临界值：空行/注释行/特殊行）
func TestNewWhoisResult(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantKey string
		wantVal string
		wantNil bool
	}{
		{
			name:    "正常行（带冒号）",
			line:    "   Domain Name: GOOGLE.COM   ",
			wantKey: "Domain Name",
			wantVal: "GOOGLE.COM",
			wantNil: false,
		},
		{
			name:    "空行",
			line:    "",
			wantNil: true,
		},
		{
			name:    "仅空格行",
			line:    "   ",
			wantNil: true,
		},
		{
			name:    "无冒号行",
			line:    "This is a test line without colon",
			wantNil: true,
		},
		{
			name:    "注释行（#开头）",
			line:    "# This is a comment",
			wantNil: true,
		},
		{
			name:    "特殊前缀行（%开头）",
			line:    "% No match found for 'INVALID.DOMAIN'",
			wantNil: true,
		},
		{
			name:    ">>>开头行",
			line:    ">>> Last update of WHOIS database: 2024-01-01 <<<",
			wantNil: true,
		},
		{
			name:    "仅冒号无值",
			line:    "Registrar:   ",
			wantKey: "Registrar",
			wantVal: "",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newWhoisResult(tt.line)
			if tt.wantNil {
				assert.Nil(t, got, "期望返回nil但未返回")
			} else {
				assert.NotNil(t, got, "期望非nil但返回nil")
				assert.Equal(t, tt.wantKey, got.Key, "Key解析错误")
				assert.Equal(t, tt.wantVal, got.Value, "Value解析错误")
			}
		})
	}
}

// TestAsnStrToNullInt 测试ASN字符串转null.Int（临界值：NA/空/非数字）
func TestAsnStrToNullInt(t *testing.T) {
	tests := []struct {
		name    string
		asnTmp  string
		wantInt int64
		wantVal bool // null.Int的Valid字段
		wantErr bool
	}{
		{
			name:    "正常数字（带空格）",
			asnTmp:  " 23028 ",
			wantInt: 23028,
			wantVal: true,
			wantErr: false,
		},
		{
			name:    "空字符串",
			asnTmp:  "",
			wantInt: 0,
			wantVal: false,
			wantErr: false,
		},
		{
			name:    "NA字符串",
			asnTmp:  " NA ",
			wantInt: 0,
			wantVal: false,
			wantErr: false,
		},
		{
			name:    "非数字字符串",
			asnTmp:  "AS23028",
			wantInt: 0,
			wantVal: false,
			wantErr: true,
		},
		{
			name:    "临界值0",
			asnTmp:  "0",
			wantInt: 0,
			wantVal: true,
			wantErr: false,
		},
		{
			name:    "超大数（合法）",
			asnTmp:  "4294967295",
			wantInt: 4294967295,
			wantVal: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := asnStrToNullInt(tt.asnTmp)
			if tt.wantErr {
				assert.Error(t, err, "期望错误但未返回")
			} else {
				assert.NoError(t, err, "非期望错误")
				assert.Equal(t, tt.wantInt, got.Int64, "Int64值错误")
				assert.Equal(t, tt.wantVal, got.Valid, "Valid字段错误")
			}
		})
	}
}

// ===================== 核心业务函数测试 =====================
// TestGetValueInWhoisResult 测试从WhoisResult取值（临界：空Result/无afterKey/不存在Key）
func TestGetValueInWhoisResult(t *testing.T) {
	// 构造测试用的WhoisResult
	testResult := &WhoisResult{
		WhoisOneResults: []*WhoisOneResult{
			{Key: "Registrar", Value: "Google LLC"},
			{Key: "Domain Name", Value: "GOOGLE.COM"},
			{Key: "Creation Date", Value: "1997-09-15"},
		},
	}

	tests := []struct {
		name     string
		result   *WhoisResult
		key      string
		afterKey string
		want     string
	}{
		{
			name:     "正常取值（无afterKey）",
			result:   testResult,
			key:      "Domain Name",
			afterKey: "",
			want:     "GOOGLE.COM",
		},
		{
			name:     "正常取值（有afterKey）",
			result:   testResult,
			key:      "Creation Date",
			afterKey: "Domain Name",
			want:     "1997-09-15",
		},
		{
			name:     "空Result",
			result:   nil,
			key:      "Domain Name",
			afterKey: "",
			want:     "",
		},
		{
			name:     "空Key",
			result:   testResult,
			key:      "",
			afterKey: "",
			want:     "",
		},
		{
			name:     "不存在的Key",
			result:   testResult,
			key:      "Invalid Key",
			afterKey: "",
			want:     "",
		},
		{
			name:     "afterKey不存在",
			result:   testResult,
			key:      "Creation Date",
			afterKey: "Invalid AfterKey",
			want:     "",
		},
		{
			name:     "大小写不敏感",
			result:   testResult,
			key:      "domain name",
			afterKey: "",
			want:     "GOOGLE.COM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetValueInWhoisResult(tt.result, tt.key, tt.afterKey)
			assert.Equal(t, tt.want, got, "取值结果不符")
		})
	}
}

// TestGetWhoisResult 测试基础whois查询（临界：空查询/无效查询）
func TestGetWhoisResult(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		// 仅验证是否返回结果，不验证具体值（因whois数据可能变化）
		wantNonEmpty bool
	}{
		{
			name:         "正常IP查询",
			query:        "8.8.8.8",
			wantErr:      false,
			wantNonEmpty: true,
		},
		{
			name:         "空查询（临界）",
			query:        "",
			wantErr:      true, // whois命令执行失败
			wantNonEmpty: false,
		},
		{
			name:         "无效查询（特殊字符）",
			query:        "@@@@@",
			wantErr:      true,
			wantNonEmpty: false,
		},
		{
			name:         "IPv6查询",
			query:        "2001:db8::1",
			wantErr:      false,
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWhoisResult(tt.query)
			if tt.wantErr {
				assert.Error(t, err, "期望错误但未返回")
				assert.Nil(t, got, "错误场景应返回nil")
			} else {
				assert.NoError(t, err, "非期望错误")
				assert.NotNil(t, got, "期望非nil但返回nil")
				if tt.wantNonEmpty {
					assert.Greater(t, len(got.WhoisOneResults), 0, "期望返回非空解析结果")
				}
			}
		})
	}
}

// TestGetWhoisResultWithConfig 测试带配置的whois查询
func TestGetWhoisResultWithConfig(t *testing.T) {
	// 使用APNIC的whois服务器查询国内IP
	config := &WhoisConfig{
		Host: "whois.apnic.net",
		Port: "43",
	}

	tests := []struct {
		name    string
		query   string
		config  *WhoisConfig
		wantErr bool
	}{
		{
			name:    "APNIC查询国内IP",
			query:   "1.1.1.1",
			config:  config,
			wantErr: false,
		},
		{
			name:    "空配置（使用默认）",
			query:   "8.8.8.8",
			config:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWhoisResultWithConfig(tt.query, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// TestWhoisAsnAddressPrefixByCymru 测试Cymru的ASN/IP查询（临界：空/无效ASN/无效IP）
func TestWhoisAsnAddressPrefixByCymru(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		wantQueryType string
		wantAsnValid  bool // ASN是否有效
		wantErr       bool
	}{
		{
			name:          "正常ASN查询",
			query:         "23028", // Team Cymru的ASN
			wantQueryType: "asn",
			wantAsnValid:  true,
			wantErr:       false,
		},
		{
			name:          "正常IP前缀查询",
			query:         "68.22.187.0/24",
			wantQueryType: "addressPrefix",
			wantAsnValid:  true,
			wantErr:       false,
		},
		{
			name:          "空查询（临界）",
			query:         "",
			wantQueryType: "",
			wantAsnValid:  false,
			wantErr:       false, // 代码返回nil,nil
		},
		{
			name:          "无效ASN（超大数）",
			query:         "9999999999",
			wantQueryType: "asn",
			wantAsnValid:  false,
			wantErr:       false,
		},
		{
			name:          "无效IP（临界值0.0.0.0）",
			query:         "0.0.0.0",
			wantQueryType: "addressPrefix",
			wantAsnValid:  false,
			wantErr:       false,
		},
		{
			name:          "非ASN/非IP查询",
			query:         "invalid-string",
			wantQueryType: "",
			wantAsnValid:  false,
			wantErr:       false,
		},
		{
			name:          "IPv6查询",
			query:         "2001:500:88:200::/56",
			wantQueryType: "addressPrefix",
			wantAsnValid:  true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WhoisAsnAddressPrefixByCymru(tt.query, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.query == "" || tt.query == "invalid-string" {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
					assert.Equal(t, tt.wantQueryType, got.QueryType, "QueryType错误")
					assert.Equal(t, tt.wantAsnValid, got.Asn.Valid, "ASN Valid字段错误")
					if tt.wantAsnValid {
						assert.Greater(t, got.Asn.Int64, int64(0), "期望有效ASN值")
					}
				}
			}
		})
	}
}

// ===================== 性能测试 =====================
// BenchmarkGetWhoisResult 基础whois查询性能测试
func BenchmarkGetWhoisResult(b *testing.B) {
	// 预热
	_, err := GetWhoisResult("8.8.8.8")
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	// 性能测试（b.N次循环）
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetWhoisResult("8.8.8.8")
	}
}

// BenchmarkWhoisAsnAddressPrefixByCymru Cymru查询性能测试
func BenchmarkWhoisAsnAddressPrefixByCymru(b *testing.B) {
	// 预热
	_, err := WhoisAsnAddressPrefixByCymru("23028", nil)
	if err != nil {
		b.Fatalf("预热失败: %v", err)
	}

	// 性能测试
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WhoisAsnAddressPrefixByCymru("23028", nil)
	}
}

// BenchmarkNewWhoisResult 单行解析性能测试
func BenchmarkNewWhoisResult(b *testing.B) {
	testLine := "Domain Name: GOOGLE.COM"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newWhoisResult(testLine)
	}
}

// //////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////
func TestGetWhoisResult1(t *testing.T) {
	q := "baidu.com"
	r, e := GetWhoisResult(q)
	fmt.Println(jsonutil.MarshalJson(r), e)

	q = "8.8.8.8"
	r, e = GetWhoisResult(q)
	fmt.Println(jsonutil.MarshalJson(r), e)

	whoisConfig := &WhoisConfig{
		Host: "whois.apnic.net",
	}
	q = "AS45090"
	r, e = GetWhoisResultWithConfig(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)
	v := GetValueInWhoisResult(r, "country", "aut-num")
	fmt.Println("country:", v)

	v = GetValueInWhoisResult(r, "source", "aut-num")
	fmt.Println("source:", v)

	v = GetValueInWhoisResult(r, "as-name", "aut-num")
	fmt.Println("as-name:", v)
}
func TestWhiosCymru1(t *testing.T) {
	host := `whois.cymru.com`
	q := `266087`
	whoisConfig := &WhoisConfig{
		Host: host,
		Port: "43",
	}
	r, e := WhoisAsnAddressPrefixByCymru(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)

	q = `216.90.108.31`
	r, e = WhoisAsnAddressPrefixByCymru(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)

	q = `216.90.0.0/16`
	r, e = WhoisAsnAddressPrefixByCymru(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)
	/*
		whois -h  whois.cymru.com AS266087
		AS Name
		Orbitel Telecomunicacoes e Informatica Ltda, BR

		whois -h  whois.cymru.com 216.90.108.31
		AS      | IP               | AS Name
		3561    | 216.90.108.31    | CENTURYLINK-LEGACY-SAVVIS, US

		whois -h  whois.cymru.com 8.0.0.0/12
		AS      | IP               | AS Name
		3356    | 8.0.0.0          | LEVEL3, US

	*/
	/*
		whois -h  whois.cymru.com "-v AS23028"
		Warning: RIPE flags used with a traditional server.
		AS      | CC | Registry | Allocated  | AS Name
		23028   | US | arin     | 2002-01-04 | TEAM-CYMRU, US
		whois -h  whois.cymru.com "-v 68.22.187.0/24"
		Warning: RIPE flags used with a traditional server.
		AS      | IP               | BGP Prefix          | CC | Registry | Allocated  | AS Name
		23028   | 68.22.187.0      | 68.22.187.0/24      | US | arin     | 2002-03-15 | TEAM-CYMRU, US
		whois -h  whois.cymru.com "-v 8.8.8.8"
		Warning: RIPE flags used with a traditional server.
		AS      | IP               | BGP Prefix          | CC | Registry | Allocated  | AS Name
		15169   | 8.8.8.8          | 8.8.8.0/24          | US | arin     | 2023-12-28 | GOOGLE, US
	*/

}
