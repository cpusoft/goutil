package whoisutil

import (
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type WhoisConfig struct {
	Host string `json:"host"` // -h whois.apnic.net
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
	WhoisOneResults []*WhoisOneResult `json:"whoisOneResults"`
}

type WhoisOneResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newWhoisResult(line string) *WhoisOneResult {
	tmp := strings.TrimSpace(line)
	belogs.Debug("newWhoisResult(): tmp:", tmp, "  len(tmp):", len(tmp), "  len(line):", len(line))
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
	c := &WhoisOneResult{
		Key:   key,
		Value: value,
	}
	belogs.Debug("GetWhoisResult(): line, tmp:", tmp, "   whoisResult:", jsonutil.MarshalJson(c))
	return c
}

/*
whois -h  whois.cymru.com "-v AS23028"
AS      | CC | Registry | Allocated  | AS Name
23028   | US | arin     | 2002-01-04 | TEAM-CYMRU, US
*/

/*
whois -h  whois.cymru.com "-v 68.22.187.0/24"
Warning: RIPE flags used with a traditional server.
AS      | IP               | BGP Prefix          | CC | Registry | Allocated  | AS Name
23028   | 68.22.187.0      | 68.22.187.0/24      | US | arin     | 2002-03-15 | TEAM-CYMRU, US
*/
type WhoisCymruResult struct {
	QueryType   string `json:"queryType"` // asn, addressprefix
	Asn         uint64 `json:"asn"`
	CountryCode string `json:"countryCode"` //cc
	Registry    string `json:"registry"`
	Allocated   string `json:"allocated"`
	Owner       string `json:"owner"` // AsName

	Ip            string `json:"ip,omitempty"`
	AddressPrefix string `json:"addressPrefix,omitempty"` // BGP Prefix

}
