package regexputil

import (
	"regexp"
)

func IsHex(s string) (bool, error) {
	return regexp.MatchString(`^[0-9a-fA-F]+$`, s)
}

func CheckPhone(phone string) bool {
	// /^\d{11}$/
	//pattern := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"
	pattern := `^\d{11}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(phone)
}

func CheckMail(mail string) bool {
	//  /^[a-z0-9]+([._\\-]*[a-z0-9])*@([a-z0-9]+[-a-z0-9]*[a-z0-9]+.){1,63}[a-z0-9]+$/;
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(mail)

}

func CheckPassword(password string) bool {
	pattern := `^.*(?=.{6,20})(?=.*\d)(?=.*[A-Za-z])(?=.*[-_=+!@#$%^&*?/]).*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(password)
}

func CheckCompany(company string) bool {
	pattern := `^[\u4e00-\u9fa5_a-zA-Z0-9_\s]{2,32}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(company)
}
