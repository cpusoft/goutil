package whoisutil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestGetWhoisResult(t *testing.T) {
	q := "baidu.com"
	r, e := GetWhoisResult(q)
	fmt.Println(jsonutil.MarshalJson(r), e)

	q = "8.8.8.8"
	r, e = GetWhoisResult(q)
	fmt.Println(jsonutil.MarshalJson(r), e)

	whoisConfig := &WhoisConfig{
		Host: "whois.apnic.net",
	}
	q = "AS45090"
	r, e = GetWhoisResultWithConfig(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)
	v := GetValueInWhoisResult(r, "country", "aut-num")
	fmt.Println("country:", v)

	v = GetValueInWhoisResult(r, "source", "aut-num")
	fmt.Println("source:", v)

	v = GetValueInWhoisResult(r, "as-name", "aut-num")
	fmt.Println("as-name:", v)
}
func TestWhiosCymru(t *testing.T) {
	host := `whois.cymru.com`
	q := `266087`
	whoisConfig := &WhoisConfig{
		Host: host,
		Port: "43",
	}
	r, e := WhoisAsnAddressPrefixByCymru(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)

	q = `216.90.108.31`
	r, e = WhoisAsnAddressPrefixByCymru(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)

	q = `216.90/16`
	r, e = WhoisAsnAddressPrefixByCymru(q, whoisConfig)
	fmt.Println(jsonutil.MarshalJson(r), e)
	/*
		whois -h  whois.cymru.com AS266087
		AS Name
		Orbitel Telecomunicacoes e Informatica Ltda, BR

		whois -h  whois.cymru.com 216.90.108.31
		AS      | IP               | AS Name
		3561    | 216.90.108.31    | CENTURYLINK-LEGACY-SAVVIS, US

		whois -h  whois.cymru.com 8.0.0.0/12
		AS      | IP               | AS Name
		3356    | 8.0.0.0          | LEVEL3, US

	*/
	/*
		whois -h  whois.cymru.com "-v AS23028"
		Warning: RIPE flags used with a traditional server.
		AS      | CC | Registry | Allocated  | AS Name
		23028   | US | arin     | 2002-01-04 | TEAM-CYMRU, US
		whois -h  whois.cymru.com "-v 68.22.187.0/24"
		Warning: RIPE flags used with a traditional server.
		AS      | IP               | BGP Prefix          | CC | Registry | Allocated  | AS Name
		23028   | 68.22.187.0      | 68.22.187.0/24      | US | arin     | 2002-03-15 | TEAM-CYMRU, US
		whois -h  whois.cymru.com "-v 8.8.8.8"
		Warning: RIPE flags used with a traditional server.
		AS      | IP               | BGP Prefix          | CC | Registry | Allocated  | AS Name
		15169   | 8.8.8.8          | 8.8.8.0/24          | US | arin     | 2023-12-28 | GOOGLE, US
	*/

}
