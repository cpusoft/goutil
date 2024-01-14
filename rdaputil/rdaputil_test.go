package rdaputil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

/*
whois.afrinic.net AS
aut-num:        AS37133
as-name:        airtel-tz-as
country:        TZ

whois.apnic.net
aut-num:        AS4134
country:        CN
as-name:        CHINANET-BACKBONE



whois -h whois.arin.net AS749 / AS577
ASNumber:       577
OrgName:        Bell Canada
Country:        CA

whois -h whois.lacnic.net
aut-num:     AS26599
owner:       TELEF NICA BRASIL S.A

whois.ripe.net
aut-num:        AS2856
as-name:        BT-UK-AS
descr:          BTnet UK Regional network
org:            ORG-BPIS1-RIPE
org-name:       British Telecommunications PLC
country:        GB


*/

// https://datatracker.ietf.org/doc/rfc7485/?include_text=1
// different items in 5 Rirs
func TestRdap(t *testing.T) {
	//	d, err := RdapDomain("baidu.com")
	//	fmt.Println(jsonutil.MarshalJson(d), err)

	apnic, err := RdapAsn(uint64(4134))
	fmt.Println("\n\napnic china:", jsonutil.MarshalJson(apnic), err)

	arin, err := RdapAsn(uint64(749))
	fmt.Println("\n\narin us:", jsonutil.MarshalJson(arin), err)

	ripe, err := RdapAsn(uint64(33915))
	fmt.Println("\n\nripe nl:", jsonutil.MarshalJson(ripe), err)

	br, err := RdapAsn(uint64(26599))
	fmt.Println("\n\nlacnic br:", jsonutil.MarshalJson(br), err)

	cl, err := RdapAsn(uint64(22047))
	fmt.Println("\n\nlacnic cl:", jsonutil.MarshalJson(cl), err)

	za, err := RdapAsn(uint64(16637))
	fmt.Println("\n\nafrinic za:", jsonutil.MarshalJson(za), err)

	//	i, err := RdapAddressPrefix("8.8.8.0/24")
	//	fmt.Println(jsonutil.MarshalJson(i), err)

}

/* APNIC
{
	"DecodeData": {
	},
	"Lang": "",
	"Conformance": [
		"history_version_0",
		"nro_rdap_profile_0",
		"nro_rdap_profile_asn_hierarchical_0",
		"cidr0",
		"rdap_level_0"
	],
	"ObjectClassName": "autnum",
	"Notices": [
		{
			"DecodeData": {
			},
			"Title": "Source",
			"Type": "",
			"Description": [
				"Objects returned came from source",
				"APNIC"
			],
			"Links": null
		},
		{
			"DecodeData": {
			},
			"Title": "Terms and Conditions",
			"Type": "",
			"Description": [
				"This is the APNIC WHOIS Database query service. The objects are in RDAP format."
			],
			"Links": [
				{
					"DecodeData": {
					},
					"Value": "https://rdap.apnic.net/autnum/4134",
					"Rel": "terms-of-service",
					"Href": "http://www.apnic.net/db/dbcopyright.html",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "text/html"
				}
			]
		},
		{
			"DecodeData": {
			},
			"Title": "Whois Inaccuracy Reporting",
			"Type": "",
			"Description": [
				"If you see inaccuracies in the results, please visit: "
			],
			"Links": [
				{
					"DecodeData": {
					},
					"Value": "https://rdap.apnic.net/autnum/4134",
					"Rel": "inaccuracy-report",
					"Href": "https://www.apnic.net/manage-ip/using-whois/abuse-and-spamming/invalid-contact-form",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "text/html"
				}
			]
		}
	],
	"Handle": "AS4134",
	"StartAutnum": 4134,
	"EndAutnum": 4134,
	"IPVersion": "",
	"Name": "CHINANET-BACKBONE",
	"Type": "",
	"Status": [
		"active"
	],
	"Country": "CN",
	"Entities": [
		{
			"DecodeData": {
			},
			"Lang": "",
			"Conformance": null,
			"ObjectClassName": "entity",
			"Notices": null,
			"Handle": "IRT-CHINANET-CN",
			"VCard": {
				"Properties": [
					{
						"Name": "version",
						"Parameters": {
						},
						"Type": "text",
						"Value": "4.0"
					},
					{
						"Name": "fn",
						"Parameters": {
						},
						"Type": "text",
						"Value": "IRT-CHINANET-CN"
					},
					{
						"Name": "kind",
						"Parameters": {
						},
						"Type": "text",
						"Value": "group"
					},
					{
						"Name": "adr",
						"Parameters": {
							"label": [
								"No.31 ,jingrong street,beijing\n100032"
							]
						},
						"Type": "text",
						"Value": [
							"",
							"",
							"",
							"",
							"",
							"",
							""
						]
					},
					{
						"Name": "email",
						"Parameters": {
						},
						"Type": "text",
						"Value": "anti-spam@chinatelecom.cn"
					},
					{
						"Name": "email",
						"Parameters": {
							"pref": [
								"1"
							]
						},
						"Type": "text",
						"Value": "anti-spam@chinatelecom.cn"
					}
				]
			},
			"Roles": [
				"abuse"
			],
			"PublicIDs": null,
			"Entities": null,
			"Remarks": [
				{
					"DecodeData": {
					},
					"Title": "remarks",
					"Type": "",
					"Description": [
						"anti-spam@chinatelecom.cn was validated on 2023-10-08"
					],
					"Links": null
				}
			],
			"Links": [
				{
					"DecodeData": {
					},
					"Value": "https://rdap.apnic.net/autnum/4134",
					"Rel": "self",
					"Href": "https://rdap.apnic.net/entity/IRT-CHINANET-CN",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "application/rdap+json"
				}
			],
			"Events": [
				{
					"DecodeData": {
					},
					"Action": "registration",
					"Actor": "",
					"Date": "2010-11-15T00:31:55Z",
					"Links": null
				},
				{
					"DecodeData": {
					},
					"Action": "last changed",
					"Actor": "",
					"Date": "2023-10-08T08:55:58Z",
					"Links": null
				}
			],
			"AsEventActor": null,
			"Status": null,
			"Port43": "",
			"Networks": null,
			"Autnums": null
		},
		{
			"DecodeData": {
			},
			"Lang": "",
			"Conformance": null,
			"ObjectClassName": "entity",
			"Notices": null,
			"Handle": "CH93-AP",
			"VCard": {
				"Properties": [
					{
						"Name": "version",
						"Parameters": {
						},
						"Type": "text",
						"Value": "4.0"
					},
					{
						"Name": "fn",
						"Parameters": {
						},
						"Type": "text",
						"Value": "Chinanet Hostmaster"
					},
					{
						"Name": "kind",
						"Parameters": {
						},
						"Type": "text",
						"Value": "individual"
					},
					{
						"Name": "adr",
						"Parameters": {
							"label": [
								"No.31 ,jingrong street,beijing\n100032"
							]
						},
						"Type": "text",
						"Value": [
							"",
							"",
							"",
							"",
							"",
							"",
							""
						]
					},
					{
						"Name": "tel",
						"Parameters": {
							"type": [
								"voice"
							]
						},
						"Type": "text",
						"Value": "+86-10-58501724"
					},
					{
						"Name": "tel",
						"Parameters": {
							"type": [
								"fax"
							]
						},
						"Type": "text",
						"Value": "+86-10-58501724"
					},
					{
						"Name": "email",
						"Parameters": {
						},
						"Type": "text",
						"Value": "anti-spam@chinatelecom.cn"
					}
				]
			},
			"Roles": [
				"technical",
				"administrative"
			],
			"PublicIDs": null,
			"Entities": null,
			"Remarks": null,
			"Links": [
				{
					"DecodeData": {
					},
					"Value": "https://rdap.apnic.net/autnum/4134",
					"Rel": "self",
					"Href": "https://rdap.apnic.net/entity/CH93-AP",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "application/rdap+json"
				}
			],
			"Events": [
				{
					"DecodeData": {
					},
					"Action": "registration",
					"Actor": "",
					"Date": "2008-09-04T07:29:13Z",
					"Links": null
				},
				{
					"DecodeData": {
					},
					"Action": "last changed",
					"Actor": "",
					"Date": "2022-02-28T06:53:44Z",
					"Links": null
				}
			],
			"AsEventActor": null,
			"Status": null,
			"Port43": "",
			"Networks": null,
			"Autnums": null
		}
	],
	"Remarks": [
		{
			"DecodeData": {
			},
			"Title": "description",
			"Type": "",
			"Description": [
				"No.31,Jin-rong Street",
				"Beijing",
				"100032"
			],
			"Links": null
		},
		{
			"DecodeData": {
			},
			"Title": "remarks",
			"Type": "",
			"Description": [
				"for backbone of chinanet"
			],
			"Links": null
		}
	],
	"Links": [
		{
			"DecodeData": {
			},
			"Value": "https://rdap.apnic.net/autnum/4134",
			"Rel": "self",
			"Href": "https://rdap.apnic.net/autnum/4134",
			"HrefLang": null,
			"Title": "",
			"Media": "",
			"Type": "application/rdap+json"
		},
		{
			"DecodeData": {
			},
			"Value": "https://rdap.apnic.net/autnum/4134",
			"Rel": "related",
			"Href": "https://netox.apnic.net/search/AS4134?utm_source=rdap\u0026utm_medium=result\u0026utm_campaign=rdap_result",
			"HrefLang": null,
			"Title": "",
			"Media": "",
			"Type": "text/html"
		}
	],
	"Port43": "whois.apnic.net",
	"Events": [
		{
			"DecodeData": {
			},
			"Action": "registration",
			"Actor": "",
			"Date": "2008-09-04T06:40:34Z",
			"Links": null
		},
		{
			"DecodeData": {
			},
			"Action": "last changed",
			"Actor": "",
			"Date": "2021-06-15T08:05:05Z",
			"Links": null
		}
	]
}

*/

/* arin
{
	"DecodeData": {
	},
	"Lang": "",
	"Conformance": [
		"nro_rdap_profile_0",
		"rdap_level_0",
		"nro_rdap_profile_asn_flat_0"
	],
	"ObjectClassName": "autnum",
	"Notices": [
		{
			"DecodeData": {
			},
			"Title": "Terms of Service",
			"Type": "",
			"Description": [
				"By using the ARIN RDAP/Whois service, you are agreeing to the RDAP/Whois Terms of Use"
			],
			"Links": [
				{
					"DecodeData": {
					},
					"Value": "https://rdap.arin.net/registry/autnum/749",
					"Rel": "terms-of-service",
					"Href": "https://www.arin.net/resources/registry/whois/tou/",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "text/html"
				}
			]
		},
		{
			"DecodeData": {
			},
			"Title": "Whois Inaccuracy Reporting",
			"Type": "",
			"Description": [
				"If you see inaccuracies in the results, please visit: "
			],
			"Links": [
				{
					"DecodeData": {
					},
					"Value": "https://rdap.arin.net/registry/autnum/749",
					"Rel": "inaccuracy-report",
					"Href": "https://www.arin.net/resources/registry/whois/inaccuracy_reporting/",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "text/html"
				}
			]
		},
		{
			"DecodeData": {
			},
			"Title": "Copyright Notice",
			"Type": "",
			"Description": [
				"Copyright 1997-2024, American Registry for Internet Numbers, Ltd."
			],
			"Links": null
		}
	],
	"Handle": "AS749",
	"StartAutnum": 749,
	"EndAutnum": 749,
	"IPVersion": "",
	"Name": "DNIC-AS-00749",
	"Type": "",
	"Status": [
		"active"
	],
	"Country": "",
	"Entities": [
		{
			"DecodeData": {
			},
			"Lang": "",
			"Conformance": null,
			"ObjectClassName": "entity",
			"Notices": null,
			"Handle": "DNIC",
			"VCard": {
				"Properties": [
					{
						"Name": "version",
						"Parameters": {
						},
						"Type": "text",
						"Value": "4.0"
					},
					{
						"Name": "fn",
						"Parameters": {
						},
						"Type": "text",
						"Value": "DoD Network Information Center"
					},
					{
						"Name": "adr",
						"Parameters": {
							"label": [
								"3990 E. Broad Street\nColumbus\nOH\n43218\nUnited States"
							]
						},
						"Type": "text",
						"Value": [
							"",
							"",
							"",
							"",
							"",
							"",
							""
						]
					},
					{
						"Name": "kind",
						"Parameters": {
						},
						"Type": "text",
						"Value": "org"
					}
				]
			},
			"Roles": [
				"registrant"
			],
			"PublicIDs": null,
			"Entities": [
				{
					"DecodeData": {
					},
					"Lang": "",
					"Conformance": null,
					"ObjectClassName": "entity",
					"Notices": null,
					"Handle": "MIL-HSTMST-ARIN",
					"VCard": {
						"Properties": [
							{
								"Name": "version",
								"Parameters": {
								},
								"Type": "text",
								"Value": "4.0"
							},
							{
								"Name": "adr",
								"Parameters": {
									"label": [
										"DISA-Columbus\n300 North James Road\nWhitehall\nOH\n43213\nUnited States"
									]
								},
								"Type": "text",
								"Value": [
									"",
									"",
									"",
									"",
									"",
									"",
									""
								]
							},
							{
								"Name": "fn",
								"Parameters": {
								},
								"Type": "text",
								"Value": "Network DoD"
							},
							{
								"Name": "org",
								"Parameters": {
								},
								"Type": "text",
								"Value": "Network DoD"
							},
							{
								"Name": "kind",
								"Parameters": {
								},
								"Type": "text",
								"Value": "group"
							},
							{
								"Name": "email",
								"Parameters": {
								},
								"Type": "text",
								"Value": "disa.columbus.ns.mbx.hostmaster-dod-nic@mail.mil"
							},
							{
								"Name": "tel",
								"Parameters": {
									"type": [
										"work",
										"voice"
									]
								},
								"Type": "text",
								"Value": "+1-844-347-2457;ext2"
							}
						]
					},
					"Roles": [
						"administrative",
						"technical"
					],
					"PublicIDs": null,
					"Entities": null,
					"Remarks": null,
					"Links": [
						{
							"DecodeData": {
							},
							"Value": "https://rdap.arin.net/registry/autnum/749",
							"Rel": "self",
							"Href": "https://rdap.arin.net/registry/entity/MIL-HSTMST-ARIN",
							"HrefLang": null,
							"Title": "",
							"Media": "",
							"Type": "application/rdap+json"
						},
						{
							"DecodeData": {
							},
							"Value": "https://rdap.arin.net/registry/autnum/749",
							"Rel": "alternate",
							"Href": "https://whois.arin.net/rest/poc/MIL-HSTMST-ARIN",
							"HrefLang": null,
							"Title": "",
							"Media": "",
							"Type": "application/xml"
						}
					],
					"Events": [
						{
							"DecodeData": {
							},
							"Action": "last changed",
							"Actor": "",
							"Date": "2023-02-10T18:17:30-05:00",
							"Links": null
						},
						{
							"DecodeData": {
							},
							"Action": "registration",
							"Actor": "",
							"Date": "1993-05-26T19:47:50-04:00",
							"Links": null
						}
					],
					"AsEventActor": null,
					"Status": [
						"validated"
					],
					"Port43": "whois.arin.net",
					"Networks": null,
					"Autnums": null
				},
				{
					"DecodeData": {
					},
					"Lang": "",
					"Conformance": null,
					"ObjectClassName": "entity",
					"Notices": null,
					"Handle": "REGIS10-ARIN",
					"VCard": {
						"Properties": [
							{
								"Name": "version",
								"Parameters": {
								},
								"Type": "text",
								"Value": "4.0"
							},
							{
								"Name": "adr",
								"Parameters": {
									"label": [
										"DISA-Columbus\n300 North James Road\nWhitehall\nOH\n43213\nUnited States"
									]
								},
								"Type": "text",
								"Value": [
									"",
									"",
									"",
									"",
									"",
									"",
									""
								]
							},
							{
								"Name": "fn",
								"Parameters": {
								},
								"Type": "text",
								"Value": "Registration"
							},
							{
								"Name": "org",
								"Parameters": {
								},
								"Type": "text",
								"Value": "Registration"
							},
							{
								"Name": "kind",
								"Parameters": {
								},
								"Type": "text",
								"Value": "group"
							},
							{
								"Name": "email",
								"Parameters": {
								},
								"Type": "text",
								"Value": "disa.columbus.ns.mbx.arin-registrations@mail.mil"
							},
							{
								"Name": "tel",
								"Parameters": {
									"type": [
										"work",
										"voice"
									]
								},
								"Type": "text",
								"Value": "+1-844-347-2457;ext2"
							}
						]
					},
					"Roles": [
						"technical",
						"abuse"
					],
					"PublicIDs": null,
					"Entities": null,
					"Remarks": null,
					"Links": [
						{
							"DecodeData": {
							},
							"Value": "https://rdap.arin.net/registry/autnum/749",
							"Rel": "self",
							"Href": "https://rdap.arin.net/registry/entity/REGIS10-ARIN",
							"HrefLang": null,
							"Title": "",
							"Media": "",
							"Type": "application/rdap+json"
						},
						{
							"DecodeData": {
							},
							"Value": "https://rdap.arin.net/registry/autnum/749",
							"Rel": "alternate",
							"Href": "https://whois.arin.net/rest/poc/REGIS10-ARIN",
							"HrefLang": null,
							"Title": "",
							"Media": "",
							"Type": "application/xml"
						}
					],
					"Events": [
						{
							"DecodeData": {
							},
							"Action": "last changed",
							"Actor": "",
							"Date": "2023-02-09T12:52:18-05:00",
							"Links": null
						},
						{
							"DecodeData": {
							},
							"Action": "registration",
							"Actor": "",
							"Date": "2009-06-24T09:41:15-04:00",
							"Links": null
						}
					],
					"AsEventActor": null,
					"Status": [
						"validated"
					],
					"Port43": "whois.arin.net",
					"Networks": null,
					"Autnums": null
				}
			],
			"Remarks": null,
			"Links": [
				{
					"DecodeData": {
					},
					"Value": "https://rdap.arin.net/registry/autnum/749",
					"Rel": "self",
					"Href": "https://rdap.arin.net/registry/entity/DNIC",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "application/rdap+json"
				},
				{
					"DecodeData": {
					},
					"Value": "https://rdap.arin.net/registry/autnum/749",
					"Rel": "alternate",
					"Href": "https://whois.arin.net/rest/org/DNIC",
					"HrefLang": null,
					"Title": "",
					"Media": "",
					"Type": "application/xml"
				}
			],
			"Events": [
				{
					"DecodeData": {
					},
					"Action": "last changed",
					"Actor": "",
					"Date": "2011-08-17T10:45:37-04:00",
					"Links": null
				}
			],
			"AsEventActor": null,
			"Status": null,
			"Port43": "whois.arin.net",
			"Networks": null,
			"Autnums": null
		}
	],
	"Remarks": null,
	"Links": [
		{
			"DecodeData": {
			},
			"Value": "https://rdap.arin.net/registry/autnum/749",
			"Rel": "self",
			"Href": "https://rdap.arin.net/registry/autnum/749",
			"HrefLang": null,
			"Title": "",
			"Media": "",
			"Type": "application/rdap+json"
		},
		{
			"DecodeData": {
			},
			"Value": "https://rdap.arin.net/registry/autnum/749",
			"Rel": "alternate",
			"Href": "https://whois.arin.net/rest/asn/AS749",
			"HrefLang": null,
			"Title": "",
			"Media": "",
			"Type": "application/xml"
		}
	],
	"Port43": "whois.arin.net",
	"Events": [
		{
			"DecodeData": {
			},
			"Action": "last changed",
			"Actor": "",
			"Date": "2011-07-05T12:17:31-04:00",
			"Links": null
		}
	]
}

*/
