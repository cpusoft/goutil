package whoisutil

import (
	"os/exec"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
)

func GetWhoisResult(q string) (results map[string]string, err error) {
	belogs.Debug("GetWhoisResult(): cmd:  whois ", q)
	cmd := exec.Command("whois", q)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("GetWhoisResult(): exec.Command: q:", q, "   err: ", err, "   output: "+string(output))
		return nil, err
	}
	result := string(output)
	tmps := strings.Split(result, osutil.GetNewLineSep())
	results = make(map[string]string, len(tmps))
	for i := range tmps {
		tmp := strings.TrimSpace(tmps[i])
		if strings.HasPrefix(tmp, "#") || strings.HasPrefix(tmp, "%") {
			continue
		}
		split := strings.Split(tmp, ":")
		key := strings.TrimSpace(split[0])
		var value string
		if len(split) > 1 {
			value = strings.TrimSpace(split[1])
		}
		if v, ok := results[key]; ok {
			v = v + " " + value
			results[key] = v
		} else {
			results[key] = value
		}
		belogs.Debug("GetWhoisResult(): key:", key, "  results[key]:", results[key])
	}
	belogs.Debug("GetWhoisResult(): len(results):", len(results))
	return results, nil
}
