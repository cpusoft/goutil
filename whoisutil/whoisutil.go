package whoisutil

import (
	"os/exec"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
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

func WhoisAsnAddressPrefixByCymru(query string,
	whoisConfig *WhoisConfig) (whoisCymruResult *WhoisCymruResult, err error) {
	belogs.Debug("WhoisAsnAddressPrefixByCymru(): query:", query, "  whoisConfig:", jsonutil.MarshalJson(whoisConfig))
	query = strings.TrimSpace(query)
	if query == "" {
		belogs.Error("WhoisAsnAddressPrefixByCymru(): query is empty")
		return nil, nil
	}
	var isQueryAsn bool
	if convert.StringIsDigit(query) {
		// asn
		query = "\"-v AS" + query + "\""
		isQueryAsn = true
	} else if strings.Contains(query, ".") || strings.Contains(query, ":") {
		// ip address or prefix
		query = "\"-v " + query + "\""
		isQueryAsn = false
	}
	belogs.Debug("WhoisAsnAddressPrefixByCymru(): new query:", query, "   isQueryAsn:", isQueryAsn)

	if whoisConfig == nil {
		whoisConfig = &WhoisConfig{
			Host: "whois.cymru.com",
			Port: "43",
		}
	}
	var cmd *exec.Cmd
	if whoisConfig == nil {
		cmd = exec.Command("whois", query)
	} else {
		cmd = exec.Command("whois", whoisConfig.getParamsWithQuery(query)...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("WhoisAsnAddressPrefixByCymru(): exec.Command fail, query:", query,
			"   output: "+string(output), err)
		return nil, err
	}
	outputStr := string(output)
	tmps := strings.Split(outputStr, osutil.GetNewLineSep())
	belogs.Debug("WhoisAsnAddressPrefixByCymru(): outputStr:", outputStr, "   tmps:", jsonutil.MarshalJson(tmps))

	whoisCymruResult = &WhoisCymruResult{}
	for i := range tmps {
		line := strings.TrimSpace(tmps[i])
		if line == "" ||
			strings.HasPrefix(line, "Warning") ||
			strings.HasPrefix(line, "AS") ||
			strings.HasPrefix(line, "Error:") {
			continue
		}
		belogs.Debug("WhoisAsnAddressPrefixByCymru(): line:", line)
		split := strings.Split(line, "|")
		if isQueryAsn {
			if len(split) != 5 {
				belogs.Error("WhoisAsnAddressPrefixByCymru(): isQueryAsn but len(slite)!=5, query:", query,
					"   line:", line, "   split:", jsonutil.MarshalJson(split))
				continue
			}
			whoisCymruResult.QueryType = "asn"
			whoisCymruResult.Asn, _ = convert.String2Uint64(strings.TrimSpace(split[0]))
			whoisCymruResult.CountryCode = strings.TrimSpace(split[1])
			whoisCymruResult.Registry = strings.TrimSpace(split[2])
			whoisCymruResult.Allocated = strings.TrimSpace(split[3])
			whoisCymruResult.Owner = strings.TrimSpace(split[4])
		} else {
			if len(split) != 7 {
				belogs.Error("WhoisAsnAddressPrefixByCymru(): isQueryIpPrefix but len(slite)!=3, query:", query,
					"   line:", line, "   split:", jsonutil.MarshalJson(split))
				continue
			}

			whoisCymruResult.QueryType = "addressPrefix"
			whoisCymruResult.Asn, _ = convert.String2Uint64(strings.TrimSpace(split[0]))
			whoisCymruResult.Ip = strings.TrimSpace(split[1])
			whoisCymruResult.AddressPrefix = strings.TrimSpace(split[2])
			whoisCymruResult.CountryCode = strings.TrimSpace(split[3])
			whoisCymruResult.Registry = strings.TrimSpace(split[4])
			whoisCymruResult.Allocated = strings.TrimSpace(split[5])
			whoisCymruResult.Owner = strings.TrimSpace(split[6])
		}
		belogs.Debug("WhoisAsnAddressPrefixByCymru(): whoisCymruResult:", jsonutil.MarshalJson(whoisCymruResult))
		break
	}
	belogs.Debug("WhoisAsnAddressPrefixByCymru(): GetWhoisResultWithConfig success, query:", query,
		"   whoisConfig:", jsonutil.MarshalJson(whoisConfig),
		"   whoisCymruResult:", jsonutil.MarshalJson(whoisCymruResult))
	return whoisCymruResult, nil
}
