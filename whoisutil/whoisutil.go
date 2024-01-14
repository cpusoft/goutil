package whoisutil

import (
	"os/exec"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func GetWhoisResult(q string) (whoisResults []*WhoisResult, err error) {
	return GetWhoisResultWithConfig(q, nil)
}
func GetWhoisResultWithConfig(q string, whoisConfig *WhoisConfig) (whoisResults []*WhoisResult, err error) {
	belogs.Debug("GetWhoisResult(): cmd:  whois ", q, "  whoisConfig:", jsonutil.MarshalJson(whoisConfig))

	var cmd *exec.Cmd
	if whoisConfig == nil {
		cmd = exec.Command("whois", q)
	} else {
		cmd = exec.Command("whois", whoisConfig.getParamsWithQuery(q)...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetWhoisResult(): exec.Command: q:", q, "   err: ", err, "   output: "+string(output))
		return nil, err
	}
	outputStr := string(output)
	tmps := strings.Split(outputStr, osutil.GetNewLineSep())
	whoisResults = make([]*WhoisResult, 0, len(tmps))
	for i := range tmps {
		whoisResult := NewWhoisResult(tmps[i])
		if whoisResults == nil {
			continue
		}
		belogs.Debug("GetWhoisResult(): line:", tmps[i], "  whoisResult:", jsonutil.MarshalJson(whoisResult))
		whoisResults = append(whoisResults, whoisResult)
	}
	belogs.Debug("GetWhoisResult(): len(whoisResults):", len(whoisResults))
	return whoisResults, nil
}
