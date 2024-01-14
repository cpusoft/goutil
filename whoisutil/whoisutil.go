package whoisutil

import (
	"os/exec"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func GetWhoisResult(q string) (whoisResult *WhoisResult, err error) {
	return GetWhoisResultWithConfig(q, nil)
}
func GetWhoisResultWithConfig(query string, whoisConfig *WhoisConfig) (whoisResult *WhoisResult, err error) {
	belogs.Debug("GetWhoisResult(): cmd:  query:", query, "  whoisConfig:", jsonutil.MarshalJson(whoisConfig))

	var cmd *exec.Cmd
	if whoisConfig == nil {
		cmd = exec.Command("whois", query)
	} else {
		cmd = exec.Command("whois", whoisConfig.getParamsWithQuery(query)...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetWhoisResult(): exec.Command fail, query:", query, "   output: "+string(output), err)
		return nil, err
	}
	outputStr := string(output)
	tmps := strings.Split(outputStr, osutil.GetNewLineSep())
	whoisOneResults := make([]*WhoisOneResult, 0, len(tmps))
	for i := range tmps {
		whoisOneResult := newWhoisResult(tmps[i])
		if whoisOneResult == nil {
			continue
		}
		belogs.Debug("GetWhoisResult(): line:", tmps[i], "  whoisOneResult:", jsonutil.MarshalJson(whoisOneResult))
		whoisOneResults = append(whoisOneResults, whoisOneResult)
	}
	whoisResult = &WhoisResult{
		WhoisOneResults: whoisOneResults,
	}
	belogs.Debug("GetWhoisResult(): whoisResult:", jsonutil.MarshalJson(whoisResult))
	return whoisResult, nil
}

// afterKey: may be "", Value is should afterKey to get by Key
func GetValueInWhoisResult(whoisResult *WhoisResult, key string, afterKey string) string {
	if whoisResult == nil || len(key) == 0 {
		return ""
	}
	k := strings.TrimSpace(key)
	aK := strings.TrimSpace(afterKey)
	var haveAfter bool
	if len(aK) == 0 {
		haveAfter = true
	}
	for i := range whoisResult.WhoisOneResults {
		if strings.EqualFold(aK, whoisResult.WhoisOneResults[i].Key) {
			haveAfter = true
		}
		if haveAfter && strings.EqualFold(k, whoisResult.WhoisOneResults[i].Key) {
			return whoisResult.WhoisOneResults[i].Value
		}
	}
	return ""
}
