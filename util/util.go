package util

import (
	"errors"
	"fmt"
	belogs "github.com/astaxie/beego/logs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func GetParentPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	dirs := strings.Split(path, string(os.PathSeparator))
	index := len(dirs)
	if len(dirs) > 2 {
		index = len(dirs) - 2
	}
	ret := strings.Join(dirs[:index], string(os.PathSeparator))
	return ret
}

// check prefix, such as 2001:DB8::/32  or 198.51.100.0/24
func CheckPrefix(prefix string) error {
	if len(prefix) == 0 {
		return nil
	}
	belogs.Info("CheckPrefix():will check prefix ", prefix)
	pos := strings.Index(prefix, "/")
	lastPos := strings.LastIndex(prefix, "/")
	if pos <= 0 || lastPos <= 0 {
		return errors.New("prefix is not contains '/' , or '/' position is error")
	}
	separatorCount := strings.Count(prefix, "/")
	if separatorCount != 1 {
		return errors.New("prefix can contain only one '/' ")
	}
	ipAndLength := strings.Split(prefix, "/")
	if len(ipAndLength) != 2 {
		return errors.New("prefix is not contains '/' or too much")
	}
	ip := ipAndLength[0]
	belogs.Info("CheckPrefix():will check by regular expression ", ip)

	//check ipv4
	patternIpv4 := `^(\d+)\.(\d+)\.(\d+)\.(\d+)$`
	matchedIpv4, errIpv4 := regexp.MatchString(patternIpv4, ip)

	//check ipv6
	patternIpv6 := `^\s*((([0-9A-Fa-f]{1,4}:){7}(([0-9A-Fa-f]{1,4})|:))|(([0-9A-Fa-f]{1,4}:){6}(:|((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})|(:[0-9A-Fa-f]{1,4})))|(([0-9A-Fa-f]{1,4}:){5}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){4}(:[0-9A-Fa-f]{1,4}){0,1}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){3}(:[0-9A-Fa-f]{1,4}){0,2}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:){2}(:[0-9A-Fa-f]{1,4}){0,3}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(([0-9A-Fa-f]{1,4}:)(:[0-9A-Fa-f]{1,4}){0,4}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(:(:[0-9A-Fa-f]{1,4}){0,5}((:((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})?)|((:[0-9A-Fa-f]{1,4}){1,2})))|(((25[0-5]|2[0-4]\d|[01]?\d{1,2})(\.(25[0-5]|2[0-4]\d|[01]?\d{1,2})){3})))(%.+)?\s*$`
	matchedIpv6, errIpv6 := regexp.MatchString(patternIpv6, ip)

	//need ipv4 or ipv6 is no err, and one check is ok
	if (errIpv4 != nil || errIpv6 != nil) &&
		matchedIpv4 == false && matchedIpv6 == false {
		return errors.New("prefix is not legal for " + ip)
	}

	//check length
	// will after check,
	return nil
}

// check asn, it is integer and should is greater than 0
// but asn's  value is zero when there is no asn, because integer default value is zero in golang.
// so when it is zero, will be ignored
func CheckAsn(asn int64) error {
	if asn == 0 {
		//ignore
	}
	if asn < 0 {
		return errors.New("asn should not be a negative number")
	}
	return nil
}

func CheckMaxPrefixLength(maxPrefixLength int64) error {
	if maxPrefixLength == 0 {
		//ignore
	}
	if maxPrefixLength < 0 {
		return errors.New("asn should not be a negative number")
	}
	return nil
}

type PrefixAndAsn struct {
	FormatPrefix    string
	PrefixLength    int64
	MaxPrefixLength int64
}

func FormatPrefix(ip string) string {
	formatIp := ""

	// format  ipv4
	ipsV4 := strings.Split(ip, ".")
	if len(ipsV4) > 1 {
		for _, ipV4 := range ipsV4 {
			ip, _ := strconv.Atoi(ipV4)
			formatIp += fmt.Sprintf("%02x", ip)
		}
		return formatIp
	}

	// format ipv6
	count := strings.Count(ip, ":")
	if count > 0 {
		count := strings.Count(ip, ":")
		if count < 7 { // total colon is 8
			needCount := 7 - count + 2 //2 is current "::", need add
			colon := strings.Repeat(":", needCount)
			ip = strings.Replace(ip, "::", colon, -1)
			belogs.Info(ip)
		}
		ipsV6 := strings.Split(ip, ":")
		belogs.Info(ipsV6)
		for _, ipV6 := range ipsV6 {
			formatIp += fmt.Sprintf("%04s", ipV6)
		}
		return formatIp
	}
	return ""
}

// check prefix and asn
func CheckPrefixAsnAndGetPrefixLength(prefix string, asn int64, maxPrefixLength int64) (PrefixAndAsn, error) {

	// cannot is empty at the same time
	if len(prefix) == 0 && asn == 0 {
		return PrefixAndAsn{}, errors.New("prefix and asn should not is empty at the same time")
	}
	prefixAndAsn := PrefixAndAsn{}
	var (
		errMsg string = ""
	)

	//check prefix

	if err := CheckPrefix(prefix); err != nil {
		errMsg += (prefix + " has error:" + err.Error() + "; ")
	}

	//check asn
	if err := CheckAsn(asn); err != nil {
		errMsg += (string(asn) + " has error:" + err.Error() + "; ")
	}

	//check maxPrefixLength
	if err := CheckMaxPrefixLength(maxPrefixLength); err != nil {
		errMsg += (string(maxPrefixLength) + " has error:" + err.Error() + "; ")
	}

	// parse prefix to get prefixLength and formatIp, and if there is no maxPrefixLength, it will equal to prefixLength.
	// such as 192.0.2.0/24, the prefixLength is 24. formatip is
	// such as 2001:DB8::/32, the prefixLength is 32. formatip is 20010DB8000000000000000000000000 . filled with 0
	//
	if len(prefix) > 0 {
		//get prefixLength and maxPrefixLength
		ipsAndLength := strings.Split(prefix, "/")
		belogs.Info("CheckPrefixAsnAndGetPrefixLength():ipsAndLength: ", ipsAndLength)

		ips := ipsAndLength[0]
		belogs.Info("ips: ", ips)

		prefixAndAsn.FormatPrefix = FormatPrefix(ips)
		belogs.Info("CheckPrefixAsnAndGetPrefixLength():FormatPrefix: ", prefixAndAsn.FormatPrefix)

		PrefixLength, err := strconv.Atoi(ipsAndLength[1])
		belogs.Info("CheckPrefixAsnAndGetPrefixLength():PrefixLength: ", PrefixLength)

		if err != nil {
			errMsg += (string(ipsAndLength[1]) + " is not a number, " + err.Error() + "; ")
		} else {
			prefixAndAsn.PrefixLength = int64(PrefixLength)
		}

		if maxPrefixLength != 0 {
			prefixAndAsn.MaxPrefixLength = maxPrefixLength
		} else {
			prefixAndAsn.MaxPrefixLength = int64(prefixAndAsn.PrefixLength)
		}

	}

	// return error
	if len(errMsg) > 0 {
		return PrefixAndAsn{}, errors.New(errMsg)
	}
	belogs.Info(fmt.Sprintf("CheckPrefixAsnAndGetPrefixLength():will return prefixAndAsn: %+v", prefixAndAsn))
	return prefixAndAsn, nil
}

func AppendRoaLocalIdSql(localIds []int) string {
	roaLocalIdsSql := "("
	for i, idTmp := range localIds {
		if i < len(localIds)-1 {
			roaLocalIdsSql += (strconv.Itoa(idTmp) + ",")
		} else {
			roaLocalIdsSql += (strconv.Itoa(idTmp))
		}
	}
	roaLocalIdsSql += ") "
	return roaLocalIdsSql
}
