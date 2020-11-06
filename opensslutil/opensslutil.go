package opensslutil

import (
	"errors"
	"os/exec"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	osutil "github.com/cpusoft/goutil/osutil"
)

func GetResultsByOpensslX509(certFile string) (results []string, err error) {
	belogs.Debug("GetResultsByOpensslX509(): cmd:  openssl", "x509", "-noout", "-text", "-in", certFile, "--inform", "der")
	cmd := exec.Command("openssl", "x509", "-noout", "-text", "-in", certFile, "--inform", "der")
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetResultsByOpensslX509(): exec.Command: certFile:", certFile, "   err: ", err, "   output: "+string(output))
		return nil, errors.New("Fail to parse by openssl, it may be not a valid x509 certificate. Error is `" + string(output) + "`. " + err.Error())
	}
	result := string(output)
	tmps := strings.Split(result, osutil.GetNewLineSep())
	results = make([]string, len(tmps))
	for i := range tmps {
		results[i] = strings.TrimSpace(tmps[i])
	}
	belogs.Debug("GetResultsByOpensslX509(): len(results):", len(results))
	return results, nil
}

func GetResultsByOpensslAns1(certFile string) (results []string, err error) {

	//https://blog.csdn.net/Zhymax/article/details/7683925
	//openssl asn1parse -in -ard.mft -inform DER
	belogs.Debug("GetResultsByOpensslAns1():cmd: openssl", "asn1parse", "-in", certFile, "--inform", "der")
	cmd := exec.Command("openssl", "asn1parse", "-in", certFile, "--inform", "der")
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetResultsByOpensslAns1(): exec.Command: certFile:", certFile, "   err: ", err, ": "+string(output))
		return nil, errors.New("Fail to parse by openssl, it may be not a valid asn1 format. Error is `" + string(output) + ". " + err.Error())
	}
	result := string(output)
	tmps := strings.Split(result, osutil.GetNewLineSep())
	results = make([]string, len(tmps))
	for i := range tmps {
		results[i] = strings.TrimSpace(tmps[i])
	}
	belogs.Debug("GetResultsByOpensslAns1(): len(results):", len(results))
	return results, nil
}
