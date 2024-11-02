package asn1addressasn

import (
	"encoding/hex"
	"fmt"
	"net"
	"testing"

	//"encoding/asn1"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"

	"github.com/cpusoft/goutil/convert"
	"github.com/stretchr/testify/assert"
)

func MakeSIA() []*SIA {
	return []*SIA{
		&SIA{
			AccessMethod: SIAManifest,
			GeneralName:  []byte("rsync://example.com/root.cer"),
		},
		&SIA{
			AccessMethod: CertRRDP,
			GeneralName:  []byte("https://example.com/notification.xml"),
		},
		&SIA{
			AccessMethod: CertRepository,
			GeneralName:  []byte("rsync://example.com/repository/"),
		},
	}
}

func MakeIPs(null bool) []IPCertificateInformation {
	if null {
		return []IPCertificateInformation{
			&IPAddressNull{
				Family: 1,
			},
		}
	}

	_, net1, _ := net.ParseCIDR("0.0.0.0/0")
	_, net2, _ := net.ParseCIDR("::/0")
	ip1 := net.ParseIP("192.168.0.1")
	ip2 := net.ParseIP("192.168.0.3")

	return []IPCertificateInformation{
		&IPNet{
			IPNet: net1,
		},
		&IPNet{
			IPNet: net2,
		},
		&IPAddressRange{
			Min: ip1,
			Max: ip2,
		},
		//&IPAddressNull{Family: 1,},
	}
}

func MakeASN(null bool) []ASNCertificateInformation {
	if null {
		return []ASNCertificateInformation{
			&ASNull{},
		}
	}
	return []ASNCertificateInformation{
		&ASNRange{
			Min: 0,
			Max: 1<<31 - 1,
		},
		&ASNRange{
			Min: 0,
			Max: 1<<31 - 1,
		},
		&ASN{
			ASN: 65001,
		},
		&ASN{
			ASN: 65002,
		},
	}
}

func TestEncodeSIA(t *testing.T) {
	sias := MakeSIA()
	siaExtension, err := EncodeSIA(sias)
	assert.Nil(t, err)

	_, err = DecodeSubjectInformationAccess(siaExtension.Value)
	assert.Nil(t, err)
}

func TestEncodeIPBlocks(t *testing.T) {
	ipBlocks := MakeIPs(true)
	ipblocksExtension, err := EncodeIPAddressBlock(ipBlocks)
	assert.Nil(t, err)
	ipblocksDec, err := DecodeIPAddressBlock(ipblocksExtension.Value)
	assert.Nil(t, err)
	assert.NotNil(t, ipblocksDec)

	ipBlocks = MakeIPs(false)
	ipblocksExtension, err = EncodeIPAddressBlock(ipBlocks)
	assert.Nil(t, err)
	ipblocksDec, err = DecodeIPAddressBlock(ipblocksExtension.Value)
	assert.Nil(t, err)
	assert.NotNil(t, ipblocksDec)
}

func TestEncodeASN(t *testing.T) {
	asns := MakeASN(true)
	asnExtension, err := EncodeASN(asns, nil)
	assert.Nil(t, err)

	asnDec, rdiDec, err := DecodeASN(asnExtension.Value)
	assert.Nil(t, err)
	assert.NotNil(t, asnDec)
	assert.NotNil(t, rdiDec)

	asns = MakeASN(false)
	asnExtension, err = EncodeASN(asns, nil)
	assert.Nil(t, err)
	asnDec, rdiDec, err = DecodeASN(asnExtension.Value)
	assert.Nil(t, err)
	assert.NotNil(t, asnDec)
	assert.NotNil(t, rdiDec)
}

func TestMakeCertificate(t *testing.T) {
	ipBlocks := MakeIPs(false)
	ipblocksExtension, err := EncodeIPAddressBlock(ipBlocks)
	assert.Nil(t, err)

	asns := MakeASN(false)
	asnExtension, err := EncodeASN(asns, nil)
	assert.Nil(t, err)

	sias := MakeSIA()
	siaExtension, err := EncodeSIA(sias)
	assert.Nil(t, err)

	cert := &x509.Certificate{
		Version:      1,
		SerialNumber: big.NewInt(42),
		Subject: pkix.Name{
			Country:      []string{"USA"},
			Organization: []string{"OctoRPKI"},
		},
		ExtraExtensions: []pkix.Extension{
			*siaExtension,
			*ipblocksExtension,
			*asnExtension,
		},
		SubjectKeyId:          []byte{1, 2, 3, 4},
		CRLDistributionPoints: []string{"https://www.example.com/crl"},
	}

	// KeyUsage!

	privkey, err := rsa.GenerateKey(rand.Reader, 1024)
	assert.Nil(t, err)
	pubkey := privkey.Public()
	_, err = x509.CreateCertificate(rand.Reader, cert, cert, pubkey, privkey)
	assert.Nil(t, err)
}

func TestASN1(t *testing.T) {
	// asn1cer5
	/*
		_, net1, _ := net.ParseCIDR("0.0.0.0/0")
		_, net2, _ := net.ParseCIDR("::/0")
		ip1 := net.ParseIP("192.168.0.1")
		ip2 := net.ParseIP("192.168.0.3")

		ips := []IPCertificateInformation{
			&IPNet{
				IPNet: net1,
			},
			&IPNet{
				IPNet: net2,
			},
			&IPAddressRange{
				Min: ip1,
				Max: ip2,
			},
		}

		for _, ip := range ips {
			ipBytes, err := ip.ASN1()
			if err != nil {
				fmt.Println(err)
			}

			asn1R := asn1.RawValue{FullBytes: ipBytes}
			fmt.Println(convert.PrintBytesOneLine(ipBytes))
			fmt.Println(asn1R)
		}
	*/
	_, net1, _ := net.ParseCIDR("10.32.0.0/12")
	_, net2, _ := net.ParseCIDR("10.64.0.0/16")
	_, net3, _ := net.ParseCIDR("10.1.0.0/16")
	_, net4, _ := net.ParseCIDR("2001:0:2::/47")
	_, net5, _ := net.ParseCIDR("2001:0:200::/39")
	_, net6, _ := net.ParseCIDR("2a05:6680::/29") // 03 05 03 2A 05 66 80
	_, net7, _ := net.ParseCIDR("2a0f:c1c0::/32") // 03 05 00 2A 0F C1 C0
	ipInfos := []IPCertificateInformation{
		&IPNet{
			IPNet: net1,
		},
		&IPNet{
			IPNet: net2,
		},
		&IPNet{
			IPNet: net3,
		},
		&IPNet{
			IPNet: net4,
		},
		&IPNet{
			IPNet: net5,
		},
		&IPNet{
			IPNet: net6,
		},
		&IPNet{
			IPNet: net7,
		},
	}
	var ipBytes []byte
	var err error
	for _, ip := range ipInfos {
		ipBytes, err = ip.ASN1()
		fmt.Println("ip, Dump:", ip, hex.Dump(ipBytes))
		fmt.Println("ip, PrintBytesOneLine:", convert.PrintBytesOneLine(ipBytes), err)
	}

	_, net8, _ := net.ParseCIDR("2a0f:c1c0::/32") // 03 05 00 2A 0F C1 C0
	ipNet8 := &IPNet{
		IPNet: net8,
	}

	ipBytes, err = ipNet8.ASN1()
	fmt.Println("ipNet8 PrintBytesOneLine:", ipBytes, convert.PrintBytesOneLine(ipBytes), err)
	fmt.Println("hex EncodeToString:", hex.EncodeToString(ipBytes))
}
