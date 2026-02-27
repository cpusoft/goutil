package xmlutil

import (
	"encoding/xml"
	"testing"
)

// 定义测试用的XML结构体（模拟业务场景）
type User struct {
	XMLName xml.Name `xml:"user"`
	ID      int      `xml:"id"`
	Name    string   `xml:"name"`
	Email   string   `xml:"email,omitempty"` // 可选字段，用于临界值测试
}

// ------------------------------ MarshalXml 测试 ------------------------------
// TestMarshalXml 测试MarshalXml的功能和临界值场景
func TestMarshalXml(t *testing.T) {
	// 表格驱动测试：覆盖正常场景、临界值、异常场景
	testCases := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool // 仅用于判断是否预期序列化失败（返回空字符串）
	}{
		{
			name:     "正常场景-完整结构体",
			input:    User{ID: 1001, Name: "张三", Email: "zhangsan@test.com"},
			expected: `<user><id>1001</id><name>张三</name><email>zhangsan@test.com</email></user>`,
			wantErr:  false,
		},
		{
			name:     "临界值-空结构体",
			input:    User{},
			expected: `<user><id>0</id><name></name></user>`,
			wantErr:  false,
		},
		{
			name:     "临界值-可选字段为空",
			input:    User{ID: 1002, Name: "李四"}, // Email为空，不序列化
			expected: `<user><id>1002</id><name>李四</name></user>`,
			wantErr:  false,
		},
		{
			name:     "临界值-输入nil",
			input:    nil,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "异常场景-循环引用（序列化失败）",
			input:    &struct{ Self *struct{} }{}, // 构造循环引用，触发xml.Marshal错误
			expected: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MarshalXml(tc.input)
			if tc.wantErr {
				if result != tc.expected {
					t.Errorf("预期返回空字符串，实际返回: %s", result)
				}
			} else {
				if result != tc.expected {
					t.Errorf("序列化结果不符\n预期: %s\n实际: %s", tc.expected, result)
				}
			}
		})
	}
}

// ------------------------------ UnmarshalXml 测试（不考虑安全） ------------------------------
// TestUnmarshalXml 测试基础反序列化功能和临界值（仅验证是否返回error，不匹配错误内容）
func TestUnmarshalXml(t *testing.T) {
	testCases := []struct {
		name     string
		inputStr string
		inputPtr interface{}
		wantErr  bool                                   // 仅判断是否预期返回error
		verifyFn func(t *testing.T, result interface{}) // 验证反序列化结果的函数
	}{
		{
			name:     "正常场景-完整XML反序列化",
			inputStr: `<user><id>1001</id><name>张三</name><email>zhangsan@test.com</email></user>`,
			inputPtr: &User{},
			wantErr:  false,
			verifyFn: func(t *testing.T, result interface{}) {
				user := result.(*User)
				if user.ID != 1001 || user.Name != "张三" || user.Email != "zhangsan@test.com" {
					t.Error("反序列化结果不匹配")
				}
			},
		},
		{
			name:     "临界值-空字符串输入",
			inputStr: "",
			inputPtr: &User{},
			wantErr:  true,
		},
		{
			name:     "临界值-nil指针输入",
			inputStr: `<user><id>1001</id></user>`,
			inputPtr: nil,
			wantErr:  true,
		},
		{
			name:     "临界值-XML标签不闭合（解析失败）",
			inputStr: `<user><id>1001</id><name>张三`, // 缺少闭合标签
			inputPtr: &User{},
			wantErr:  true,
		},
		{
			name:     "临界值-部分字段缺失（兼容解析）",
			inputStr: `<user><id>1002</id></user>`, // 缺少Name/Email
			inputPtr: &User{},
			wantErr:  false,
			verifyFn: func(t *testing.T, result interface{}) {
				user := result.(*User)
				if user.ID != 1002 || user.Name != "" {
					t.Error("部分字段缺失时反序列化结果异常")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := UnmarshalXml(tc.inputStr, tc.inputPtr)
			// 仅验证是否返回error，不匹配错误内容
			if tc.wantErr {
				if err == nil {
					t.Error("预期返回error，但实际未返回")
				}
			} else {
				if err != nil {
					t.Errorf("预期不返回error，但实际返回: %v", err)
				} else if tc.verifyFn != nil {
					tc.verifyFn(t, tc.inputPtr)
				}
			}
		})
	}
}

// ------------------------------ UnmarshalXmlStrict 测试（含安全） ------------------------------
// TestUnmarshalXmlStrict 测试严格模式反序列化（仅验证是否返回error，不匹配错误内容）
func TestUnmarshalXmlStrict(t *testing.T) {
	testCases := []struct {
		name     string
		inputStr string
		inputPtr interface{}
		wantErr  bool // 仅判断是否预期返回error
		verifyFn func(t *testing.T, result interface{})
	}{
		{
			name:     "正常场景-合法UTF-8 XML",
			inputStr: `<user><id>1001</id><name>张三</name></user>`,
			inputPtr: &User{},
			wantErr:  false,
			verifyFn: func(t *testing.T, result interface{}) {
				user := result.(*User)
				if user.ID != 1001 || user.Name != "张三" {
					t.Error("严格模式反序列化结果不匹配")
				}
			},
		},
		{
			name: "安全测试-XXE外部实体（拒绝解析）",
			inputStr: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE root [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>
<user><id>1001</id><name>&xxe;</name></user>`,
			inputPtr: &User{},
			wantErr:  true,
		},
		{
			name: "安全测试-非UTF-8字符集（拒绝解析）",
			inputStr: `<?xml version="1.0" encoding="GBK"?>
<user><id>1001</id><name>张三</name></user>`,
			inputPtr: &User{},
			wantErr:  true,
		},
		{
			name:     "临界值-空字符串输入",
			inputStr: "",
			inputPtr: &User{},
			wantErr:  true,
		},
		{
			name:     "临界值-nil指针输入",
			inputStr: `<user><id>1001</id></user>`,
			inputPtr: nil,
			wantErr:  true,
		},
		{
			name:     "临界值-XML语法不规范（严格模式拒绝）",
			inputStr: `<user><id>1001</id><name>张三`, // 缺少闭合标签
			inputPtr: &User{},
			wantErr:  true,
		},
		{
			name:     "临界值-非HTML内置实体（拒绝解析）",
			inputStr: `<user><id>1001</id><name>&nonhtmlentity;</name></user>`, // 非内置实体
			inputPtr: &User{},
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := UnmarshalXmlStrict(tc.inputStr, tc.inputPtr)
			// 仅验证是否返回error，不匹配错误内容
			if tc.wantErr {
				if err == nil {
					t.Error("预期返回error，但实际未返回")
				}
			} else {
				if err != nil {
					t.Errorf("预期不返回error，但实际返回: %v", err)
				} else if tc.verifyFn != nil {
					tc.verifyFn(t, tc.inputPtr)
				}
			}
		})
	}
}

// ------------------------------ 性能测试 ------------------------------
// BenchmarkMarshalXml 测试MarshalXml序列化性能
func BenchmarkMarshalXml(b *testing.B) {
	// 准备测试数据（避免循环内重复创建）
	testData := User{ID: 1001, Name: "张三", Email: "zhangsan@test.com"}
	b.ResetTimer() // 重置计时器，排除数据准备耗时

	for i := 0; i < b.N; i++ {
		MarshalXml(testData)
	}
}

// BenchmarkUnmarshalXml 测试基础反序列化性能
func BenchmarkUnmarshalXml(b *testing.B) {
	testXml := `<user><id>1001</id><name>张三</name><email>zhangsan@test.com</email></user>`
	result := &User{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = UnmarshalXml(testXml, result)
	}
}

// BenchmarkUnmarshalXmlStrict 测试严格模式反序列化性能
func BenchmarkUnmarshalXmlStrict(b *testing.B) {
	testXml := `<user><id>1001</id><name>张三</name><email>zhangsan@test.com</email></user>`
	result := &User{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = UnmarshalXmlStrict(testXml, result)
	}
}
