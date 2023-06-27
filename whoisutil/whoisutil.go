package whoisutil

import (
	"os/exec"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
)

func GetWhoisResult(q string) (results []string, err error) {
	belogs.Debug("GetWhoisResult(): cmd:  whois ", q)

	cmd := exec.Command("whois", q)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetWhoisResult(): exec.Command: q:", q, "   err: ", err, "   output: "+string(output))
		return nil, err
	}
	outputStr := string(output)
	tmps := strings.Split(outputStr, osutil.GetNewLineSep())
	results = make([]string, 0, len(tmps))
	for i := range tmps {
		tmp := strings.TrimSpace(tmps[i])
		// >>> Last update of whois database: 2022-11-24T12:49:15Z <<<
		// For more information on Whois status codes, please visit https://icann.org/epp
		if strings.HasPrefix(tmp, ">>>") {
			break
		}
		if len(tmp) == 0 || strings.HasPrefix(tmp, "#") ||
			strings.HasPrefix(tmp, "%") || strings.HasPrefix(tmp, "No match found for") {
			continue
		}

		split := strings.SplitN(tmp, ":", 2)
		key := strings.TrimSpace(split[0])
		var value string
		if len(split) > 1 {
			value = strings.TrimSpace(split[1])
		}
		line := key + ":" + value
		results = append(results, line)
		belogs.Debug("GetWhoisResult(): line:", line)
	}
	belogs.Debug("GetWhoisResult(): len(results):", len(results))
	return results, nil
}
