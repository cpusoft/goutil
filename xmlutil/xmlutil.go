package xmlutil

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/cpusoft/goutil/belogs"
)

// str := MarshalXml(user)
func MarshalXml(f interface{}) string {
	if f == nil {
		return ""
	}

	body, err := xml.Marshal(f)
	if err != nil {
		belogs.Error("MarshalXml() failed: ", err)
		return ""
	}
	return string(body)
}

/*
var user1 = User{}
UnmarshalXml(body1, &user1)
*/
func UnmarshalXml(str string, f interface{}) error {
	// 基础输入校验
	if str == "" {
		return errors.New("unmarshal xml failed: input string is empty")
	}
	if f == nil {
		return errors.New("unmarshal xml failed: destination pointer is nil")
	}
	return xml.Unmarshal([]byte(str), f)
}

func UnmarshalXmlStrict(str string, f interface{}) error {

	// 基础输入校验
	if str == "" {
		return errors.New("unmarshal xml failed: input string is empty")
	}
	if f == nil {
		return errors.New("unmarshal xml failed: destination pointer is nil")
	}

	// 显式配置 XML 解析器，禁用外部实体（加固 XXE 防护）
	decoder := xml.NewDecoder(strings.NewReader(str))
	decoder.Strict = true           // 严格模式，拒绝不符合 XML 规范的内容
	decoder.Entity = xml.HTMLEntity // 仅允许 HTML 内置实体，禁用外部实体
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		// 限制字符集为 UTF-8，防止字符集注入风险
		if charset != "UTF-8" && charset != "utf-8" {
			return nil, errors.New("unsupported charset: " + charset)
		}
		return input, nil
	}

	err := decoder.Decode(f)
	if err != nil {
		belogs.Error("UnmarshalXmlStrict() failed, str:", str, err)
		return errors.New("unmarshal xml failed")
	}
	return nil
}
