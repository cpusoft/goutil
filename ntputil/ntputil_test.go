package ntputil

import (
	"fmt"
	"testing"
)

func TestGetNtpTim(t *testing.T) {
	//tm, err := GetNtpTime()
	//fmt.Println(tm, err)

	s, _ := GetFormatNtpTime("")
	fmt.Println("GetFormatNtpTime", s)
}
