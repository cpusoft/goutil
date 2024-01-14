package whoisutil

import (
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type WhoisConfig struct {
	Host string `json:"hots"` // -h whois.apnic.net
	Port string `json:"port"` // default: 43
}

func (c *WhoisConfig) getParamsWithQuery(query string) []string {
	params := make([]string, 0)
	params = append(params, query)
	if len(c.Host) > 0 {
		params = append(params, "-h")
		params = append(params, c.Host)
	}
	if len(c.Port) > 0 {
		params = append(params, "-p")
		params = append(params, c.Port)
	}
	belogs.Debug("WhoisConfig.getParamsWithQuery(): query:", query, "  whoisConfig:", jsonutil.MarshalJson(c),
		"   params:", jsonutil.MarshalJson(params))
	return params
}

type WhoisResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewWhoisResult(line string) *WhoisResult {
	belogs.Debug("NewWhoisResult(): line:", line)
	tmp := strings.TrimSpace(line)

	if len(tmp) == 0 ||
		!strings.Contains(tmp, ":") ||
		strings.HasPrefix(tmp, ">>>") ||
		strings.HasPrefix(tmp, "#") ||
		strings.HasPrefix(tmp, "%") ||
		strings.HasPrefix(tmp, "No match found for") ||
		strings.HasPrefix(tmp, "For more information on") {
		return nil
	}

	split := strings.SplitN(tmp, ":", 2)
	key := strings.TrimSpace(split[0])
	var value string
	if len(split) > 1 {
		value = strings.TrimSpace(split[1])
	}
	c := &WhoisResult{
		Key:   key,
		Value: value,
	}
	belogs.Debug("GetWhoisResult(): line:", line, "   whoisResult:", jsonutil.MarshalJson(c))
	return c
}
