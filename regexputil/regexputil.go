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
	// 修复邮箱正则：
	// 1. 域名段允许单字符（([0-9a-z]([0-9a-z-]{0,30}[0-9a-z])?\.)）
	// 2. 用户名长度限制为1-31位（{0,30}），32位则失败
	mailRegex = regexp.MustCompile(`^[0-9a-z][_.0-9a-z-]{0,30}@([0-9a-z]([0-9a-z-]{0,30}[0-9a-z])?\.){1,4}[a-z]{2,4}$`)
	// 拆分密码校验的正则（无前瞻断言，分步校验）
	passwordHasDigit   = regexp.MustCompile(`\d`)               // 包含数字
	passwordHasLetter  = regexp.MustCompile(`[A-Za-z]`)         // 包含字母
	passwordHasSpecial = regexp.MustCompile(`[-_=+!@#$%^&*?/]`) // 包含指定特殊字符
	// 匹配公司名称（2-32位，包含中文、字母、数字、下划线、空白符）
	companyRegex = regexp.MustCompile(`^[\p{Han}a-zA-Z0-9_\s]{2,32}$`)
)

func IsHex(s string) (bool, error) {
	return hexRegex.MatchString(s), nil
}

func CheckRpkiFileName(s string) bool {
	return rpkiFileNameRegex.MatchString(s)
}

func CheckPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

func CheckMail(mail string) bool {
	return mailRegex.MatchString(mail)
}

func CheckPassword(password string) bool {
	// 步骤1：校验长度（6-20位）
	length := len(password)
	if length < 6 || length > 20 {
		return false
	}
	// 步骤2：校验包含数字
	if !passwordHasDigit.MatchString(password) {
		return false
	}
	// 步骤3：校验包含字母
	if !passwordHasLetter.MatchString(password) {
		return false
	}
	// 步骤4：校验包含指定特殊字符
	if !passwordHasSpecial.MatchString(password) {
		return false
	}
	// 所有条件满足
	return true
}

func CheckCompany(company string) bool {
	return companyRegex.MatchString(company)
}
