package asn1cert

import (
	"fmt"
	"net"
	"testing"

	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseToAddressPrefix(t *testing.T) {
	files := []string{
		`1.roa`,
		`asn0.roa`,
		`ok.roa`,
		`fail1.roa`,
	}

	for _, file := range files {
		fmt.Println("file:", file)
		b, err := fileutil.ReadFileToBytes(file)
		if err != nil {
			fmt.Println("!!!!!!ReadFileToBytes fail:", file, err)
			continue
		}

		contentInfo := ContentInfo{}
		_, err = Unmarshal(b, &contentInfo)
		if err != nil {
			fmt.Println("!!!!!!Unmarshal contentInfo, file:", file, err)
			continue
		}
		contentTypeOid := contentInfo.ContentType.String()
		fmt.Println("file contentTypeOid:", file, contentTypeOid)

		roaSignedData := RoaSignedData{}
		for _, seq := range contentInfo.Seqs {
			//fmt.Println("seq:", jsonutil.MarshallJsonIndent(seq))

			if seq.Class == 0 && seq.Tag == 2 && !seq.IsCompound {
				// version:       version CMSVersion INTEGER 3
				roaSignedData.Version = convert.Bytes2Uint64(seq.Bytes)
			} else if seq.Class == 0 && seq.Tag == 17 && seq.IsCompound && len(seq.Bytes) < 100 {
				// digestAlgorithms DigestAlgorithmIdentifiers or signerInfos SignerInfos SET (1 elem)
				var algorithmIdentifier AlgorithmIdentifier
				_, err = Unmarshal(seq.Bytes, &algorithmIdentifier)
				if err != nil {
					fmt.Println("!!!!!!algorithmIdentifier fail:", err)
					continue
				}
				roaSignedData.AlgorithmIdentifier = algorithmIdentifier.Algorithm.String()

			} else if seq.Class == 0 && seq.Tag == 16 && seq.IsCompound {
				//  encapContentInfo EncapsulatedContentInfo
				var roaOctetString RoaOctetString
				_, err = Unmarshal(seq.FullBytes, &roaOctetString)
				if err != nil {
					fmt.Println("!!!!!Unmarshal roaOctetString fail:", err)
					continue
				}
				//fmt.Println("roaOctetString, EContentType:", roaOctetString.EContentType, " len(OctetString):", len(roaOctetString.OctetString))

				routeOriginAttestation := RouteOriginAttestation{}
				_, err = Unmarshal([]byte(roaOctetString.OctetString), &routeOriginAttestation)
				if err != nil {
					fmt.Println("!!!!!!Unmarshal routeOriginAttestation fail:", err)
					continue
				}
				//fmt.Println("RoaOctetString: routeOriginAttestation", jsonutil.MarshalJson(routeOriginAttestation))

				roaSignedData.RoaModel.Asn = int64(routeOriginAttestation.AsID)
				//roaIpAddressModels := make([]RoaIpAddressModel, 0)
				for i := range routeOriginAttestation.IpAddrBlocks {
					ipAddrBlock := routeOriginAttestation.IpAddrBlocks[i]
					addressFamily := convert.BytesToBigInt(ipAddrBlock.AddressFamily)
					//fmt.Println("addressFamily:", addressFamily)
					var size int
					if addressFamily.Uint64() == 1 {
						size = 4
					} else if addressFamily.Uint64() == 2 {
						size = 16
					}

					for j := range ipAddrBlock.Addresses {

						ipAddr := make([]byte, size)
						copy(ipAddr, ipAddrBlock.Addresses[j].Address.Bytes)
						mask := net.CIDRMask(ipAddrBlock.Addresses[j].Address.BitLength, size*8)
						//fmt.Println("ipAddr:", convert.PrintBytesOneLine(ipAddr),	jsonutil.MarshalJson(ipAddr),"  mask:", mask)
						ipNet := net.IPNet{
							IP:   net.IP(ipAddr),
							Mask: mask,
						}
						maxlength := ipAddrBlock.Addresses[j].MaxLength
						roaPrefixAddress := ipNet.String()
						//fmt.Println("file:roaPrefixAddress:", file, roaPrefixAddress, "  maxlength:", maxlength)
						fmt.Sprintf("%s,%d", roaPrefixAddress, maxlength)
					}
					fmt.Println("file len(ipAddrBlock.Addresses):", file, len(ipAddrBlock.Addresses))
				}
				//fmt.Println("roaIpAddressModels:", jsonutil.MarshalJson(roaIpAddressModels))
			}
		}
		fmt.Println("file roaSignedData:", file, jsonutil.MarshalJson(roaSignedData))
		fmt.Println("\n\n\n\n")
	}

}
