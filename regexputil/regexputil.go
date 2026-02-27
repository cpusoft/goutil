package regexputil

import (
	"regexp"
)

// 预编译所有正则表达式（包级别变量，仅初始化一次）
var (
	// 匹配十六进制字符串
	hexRegex = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	// 匹配RPKI文件名（修正-的位置，避免范围解析歧义）
	rpkiFileNameRegex = regexp.MustCompile(`^[0-9a-zA-Z_-]+\.(cer|roa|crl|mft|gbr|asa|sig|moa|toa)$`)
	// 匹配11位数字手机号
	phoneRegex = regexp.MustCompile(`^\d{11}$`)
	// 匹配邮箱格式
	mailRegex = regexp.MustCompile(`^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`)
	// 匹配密码格式（6-20位，包含数字、字母、特殊字符）
	passwordRegex = regexp.MustCompile(`^.*(?=.{6,20})(?=.*\d)(?=.*[A-Za-z])(?=.*[-_=+!@#$%^&*?/]).*$`)
	// 匹配公司名称（2-32位，包含中文、字母、数字、下划线、空白符）
	companyRegex = regexp.MustCompile(`^[\u4e00-\u9fa5_a-zA-Z0-9_\s]{2,32}$`)
)

func IsHex(s string) (bool, error) {
	// 预编译后仅调用MatchString，error恒为nil（保持原返回值结构）
	return hexRegex.MatchString(s), nil
}

// https://www.iana.org/assignments/rpki/rpki.xhtml
func CheckRpkiFileName(s string) bool {
	// 使用预编译的正则，避免重复编译
	return rpkiFileNameRegex.MatchString(s)
}

func CheckPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func CheckMail(mail string) bool {
	return mailRegex.MatchString(mail)
}

func CheckPassword(password string) bool {
	return passwordRegex.MatchString(password)
}

func CheckCompany(company string) bool {
	return companyRegex.MatchString(company)
}
