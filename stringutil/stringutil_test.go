package stringutil

import (
	"fmt"
	"strings"
	"testing"
)

func TestTrimSpaceAndNewLine(t *testing.T) {
	s := `MIIL1gYJKoZIhvcNAQcCoIILxzC   
	CC8MCAQMxDTALBglghkgBZQMEAgEwggMXBgsqhkiG9w0BCRABGqCCAwYEggMCMIIC/gICBd0YDzIwMjMxMjIwMTkyOTAwWhgPMjA
	yMzEyMjExOTM0MDBaBglghkgBZQMEAgEwggLJMGcWQjMyMzYzMDMyM2E2NjY1NjQ2MTNhNjQzNTM4M2EzYTJmMzQzODJkMzQzODI
	wM2QzZTIwMzEzNDMxMzczMTMyLnJvYQMhAFicfryQK0kBExiMpBKZBiSpDqGqYm3INk9U4NhBqBQgMFEWLDczRENFRUMyNUIzM0U
	xQjAwREYyOEQ3RDczQkFCNkQwOEI5Q0FGRkMuY3JsAyEA2gvpQ8nuGDkFNZ1pkVJOYWMk1VliGGHlcgzqdBCVZq0wZxZCMzIzNjM
	   2q0eTI87VZv4CfABTjQ==`
	s1 := strings.Replace(s, "\r", "", -1)
	s1 = strings.Replace(s1, "\n", "", -1)
	s1 = strings.Replace(s1, "\t", "", -1)
	s1 = strings.Replace(s1, " ", "", -1)
	fmt.Println("s1:", s1)

	s2 := TrimSpaceAndNewLine(s)
	fmt.Println("s2:", s2)
}
func TestTrimeSuffixAll(t *testing.T) {
	ips := []string{"16.70.0.0", "16.0.1.0"}

	for _, ip := range ips {
		str := TrimeSuffixAll(ip, ".0")
		fmt.Println(ip, " --> ", str)

	}
}

func TestGetValueFromJointStr(t *testing.T) {
	str := `a=111&b=222&c=333`
	v := GetValueFromJointStr(str, "a", "&")
	fmt.Println(v)
	v = GetValueFromJointStr(str, "b", "&")
	fmt.Println(v)
	v = GetValueFromJointStr(str, "c", "&")
	fmt.Println(v)
}

func TestOmitString(t *testing.T) {
	str := `0123456789a`
	str1 := OmitString(str, 0)
	fmt.Println(str1)
	str1 = OmitString(str, 1)
	fmt.Println(str1)
	str1 = OmitString(str, 9)
	fmt.Println(str1)
	str1 = OmitString(str, 10)
	fmt.Println(str1)
	str1 = OmitString(str, 11)
	fmt.Println(str1)
	str1 = OmitString(str, 12)
	fmt.Println(str1)
}
