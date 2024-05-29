package rrdputil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xmlutil"
)

func TestRrdpDelta(t *testing.T) {
	body := `<delta xmlns="http://www.ripe.net/rpki/rrdp" version="1" session_id="b14bf9f4-2d4f-418f-8f47-819710b78724" serial="3024">
<publish uri="rsync://rpki.luys.cloud/repo/LY-RPKI/0/73DCEEC25B33E1B00DF28D7D73BAB6D08B9CAFFC.mft" hash="543744919260b2ea96a81b2e5d76ff452da30ed520d469af39909529df0944cd">MIIL1gYJKoZIhvcNAQcCoIILxzCCC8MCAQMxDTALBglghkgBZQMEAgEwggMXBgsqhkiG9w0BCRABGqCCAwYEggMCMIIC/gICBd0YDzIwMjMxMjIwMTkyOTAwWhgPMjAyMzEyMjExOTM0MDBaBglghkgBZQMEAgEwggLJMGcWQjMyMzYzMDMyM2E2NjY1NjQ2MTNhNjQzNTM4M2EzYTJmMzQzODJkMzQzODIwM2QzZTIwMzEzNDMxMzczMTMyLnJvYQMhAFicfryQK0kBExiMpBKZBiSpDqGqYm3INk9U4NhBqBQgMFEWLDczRENFRUMyNUIzM0UxQjAwREYyOEQ3RDczQkFCNkQwOEI5Q0FGRkMuY3JsAyEA2gvpQ8nuGDkFNZ1pkVJOYWMk1VliGGHlcgzqdBCVZq0wZxZCMzIzNjMwMzIzYTY2NjU2NDYxM2E2NDM1MzYzYTNhMmYzNDM4MmQzNDM4MjAzZDNlMjAzMTM0MzEzNzMxMzIucm9hAyEAZqcux++skmateywJPnlg1mqAGktM0RZwuQ5t79dRQuUwZxZCMzIzNjMwMzIzYTY2NjU2NDYxM2E2NDM1MzEzYTNhMmYzNDM4MmQzNDM4MjAzZDNlMjAzMTM0MzEzNzMxMzIucm9hAyEAck9q+UEopLzRL/R2eQavJ4J+vwLGF0LAaDTuR7MNpQkwZxZCMzIzNjMwMzIzYTY2NjU2NDYxM2E2NDM1MzczYTNhMmYzNDM4MmQzNDM4MjAzZDNlMjAzMTM0MzEzNzMxMzIucm9hAyEAOg6vFATyxoFCYOhsc2Ed2GthrVSmtZV0A0dBX9BpDtswZxZCMzIzNjMwMzIzYTY2NjU2NDYxM2E2NDM1MzUzYTNhMmYzNDM4MmQzNDM4MjAzZDNlMjAzMTM0MzEzNzMxMzIucm9hAyEAh1zmcV+dwpKdB4vxk30wQWlQ7j+UQrVnnGgc73zDjBAwZxZCMzIzNjMwMzIzYTY2NjU2NDYxM2E2NDM1MzkzYTNhMmYzNDM4MmQzNDM4MjAzZDNlMjAzMTM0MzEzNzMxMzIucm9hAyEAB1QJC79G00TaE3hTtKlvGdpqk6YA5R8mM9CSr9U0OXWgggbkMIIG4DCCBcigAwIBAgIUQFwC8ps1tYdJ9K1WCjXHd39XmMswDQYJKoZIhvcNAQELBQAwMzExMC8GA1UEAxMoNzNEQ0VFQzI1QjMzRTFCMDBERjI4RDdENzNCQUI2RDA4QjlDQUZGQzAeFw0yMzEyMjAxOTI5MDBaFw0yMzEyMjcxOTM0MDBaMIICLTGCAikwggIlBgNVBAMTggIcMzA4MjAxMEEwMjgyMDEwMTAwQ0I1NzE5RTlCQ0VBRjcxRDExRDBFQzlBQzAwNzNGMDczMzQ0NkJBODcyNUE1NUIyMzlFQzQ5QzgyRkJBNTcwNjg4QUM5RjYwNjZGRTY3NURDNEVDNTc2NzkxNTRDNUZCNjBBREZDOEE0RURFRUI4OUEyN0RDQzBFMTc3MkJDOUZBRkVFNEQyQUQ5MzU1OUNFNkEyNzk5RDQ5NjFERUFBNkJBODkyRDUyQjlFQkZCMUI4NzZGMzc4QUQ2MEEyN0E2RkVCNDA1NkJEM0JEQTBENjRCMkYzQkY2MUFGODBCNjQ2M0NDOEQ5MEVCM0M3QUEzQTBCQzMyODlEODY2MjMwQUZGRDJDMzYzNUZCQjlEN0ZBNjY2NENBM0ZFMjRGNEU5RUJERUIyNjg3MEVFOEJGODNFNERGMTFDNDI2MkRGNUFDNkY5QkZEMzYxMjFFRDM4NzE3RkU2RUUwNEExODlCRTZCQUREQzZDOTI4MURBNDNGRkE4MzBCNTI0Mjc3QUQ4MjEzNjE4N0E2Q0JDRUUzNDdDNTVGRTBDRTQzOURDNzFBMkFBNzJFRTU0OUZGRDQyQUU4ODE5RTY0NzJEOTlFNzg4QjJGNEJDOUNGNEFFOEZCOUQzNjdDRkVBMDU3NUQ5QkRCMDlDRDNEODUxQjZCRTAxNUJENUNCMURFQzdBMEIwMjAzMDEwMDAxMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAy1cZ6bzq9x0R0OyawAc/BzNEa6hyWlWyOexJyC+6VwaIrJ9gZv5nXcTsV2eRVMX7YK38ik7e64mifcwOF3K8n6/uTSrZNVnOaieZ1JYd6qa6iS1Suev7G4dvN4rWCiem/rQFa9O9oNZLLzv2GvgLZGPMjZDrPHqjoLwyidhmIwr/0sNjX7udf6ZmTKP+JPTp696yaHDui/g+TfEcQmLfWsb5v9NhIe04cX/m7gShib5rrdxskoHaQ/+oMLUkJ3rYITYYemy87jR8Vf4M5DnccaKqcu5Un/1CrogZ5kctmeeIsvS8nPSuj7nTZ8/qBXXZvbCc09hRtr4BW9XLHex6CwIDAQABo4IB7jCCAeowHQYDVR0OBBYEFHdzHrYVJBn4IGYanVtEsLhtlaHUMB8GA1UdIwQYMBaAFHPc7sJbM+GwDfKNfXO6ttCLnK/8MA4GA1UdDwEB/wQEAwIHgDBkBgNVHR8EXTBbMFmgV6BVhlNyc3luYzovL3Jwa2kubHV5cy5jbG91ZC9yZXBvL0xZLVJQS0kvMC83M0RDRUVDMjVCMzNFMUIwMERGMjhEN0Q3M0JBQjZEMDhCOUNBRkZDLmNybDBtBggrBgEFBQcBAQRhMF8wXQYIKwYBBQUHMAKGUXJzeW5jOi8vc2FrdXlhLm5hdC5tb2UvcmVwby9OQVRPQ0EvMC83M0RDRUVDMjVCMzNFMUIwMERGMjhEN0Q3M0JBQjZEMDhCOUNBRkZDLmNlcjBvBggrBgEFBQcBCwRjMGEwXwYIKwYBBQUHMAuGU3JzeW5jOi8vcnBraS5sdXlzLmNsb3VkL3JlcG8vTFktUlBLSS8wLzczRENFRUMyNUIzM0UxQjAwREYyOEQ3RDczQkFCNkQwOEI5Q0FGRkMubWZ0MBgGA1UdIAEB/wQOMAwwCgYIKwYBBQUHDgIwIQYIKwYBBQUHAQcBAf8EEjAQMAYEAgABBQAwBgQCAAIFADAVBggrBgEFBQcBCAEB/wQGMASgAgUAMA0GCSqGSIb3DQEBCwUAA4IBAQAS0vNa6ffmQpeGKK7jEYW8oj4YxDUYp8NTDFd5NbilNQuNqz8fP/OzaXBGmYqdWSsxRFsrVqvG1HLmBivPQD6tNuzlrh3CadHPksIRKzkwAD8yMV5CCQP4MfFUwPiPGCMxt8U+WQZYA4WZz/J4jV6ENg/g3zGudmnQDEJS2ITakRIIDLv0iauUGTfIaBf1nzA3sEKueUpPN78Lv6/CNeIz7O+Ss9BPX6vdphgDQWwULaX4XR3RNmLvHfiPYXV8r/9O8OyZRgBicRx8Uym8KeZ2jbbQc33CV/6RAjA+NsBJG/o6Hr9m7pY0iA4ntbCvMIFBjMJIrL2rYUM4h11CgwR4MYIBqjCCAaYCAQOAFHdzHrYVJBn4IGYanVtEsLhtlaHUMAsGCWCGSAFlAwQCAaBrMBoGCSqGSIb3DQEJAzENBgsqhkiG9w0BCRABGjAcBgkqhkiG9w0BCQUxDxcNMjMxMjIwMTkzNDAwWjAvBgkqhkiG9w0BCQQxIgQgE5UPDDvCRSMI1TSCZFhkQCIxyucZ+4ktYfZTwry+ghIwDQYJKoZIhvcNAQEBBQAEggEArQOdBcdfGniqom4rLiMiBTI7/e16ImbgUpr/Sk7uBEuHR52RqXZpRHlf1p7/vZ9dHzJ2X2UmHn3rum2Smco2cL4EQ/iKUPGUZw3PpaH8PkyPlpa4jkMSBqE27wtT9IzPkYTn/lpwxqAffPSk7jU8LhvvVGkDqm6Szyo8eON7S+WNKOHM4pn4R7Zp5HEqTa4HWEVV9mTDrDFPXkmseEUYXTHRsLow0rkUh2uiSCGb6mFnLQ/ADeachlpAduyB9ViNjz9+e33n1T1L/IXksWsQj80W1pKEWuQse9iuiI+nPeEUJ97cA2ROjdN70UEx/jRwZYK2q0eTI87VZv4CfABTjQ==</publish>
<publish uri="rsync://rpki.luys.cloud/repo/LY-RPKI/0/73DCEEC25B33E1B00DF28D7D73BAB6D08B9CAFFC.crl" hash="bfce6605cf104ffe3a54f9e75e2020224256f0780c71ef31dd212087cc0bd403">MIIB1zCBwAIBATANBgkqhkiG9w0BAQsFADAzMTEwLwYDVQQDEyg3M0RDRUVDMjVCMzNFMUIwMERGMjhEN0Q3M0JBQjZEMDhCOUNBRkZDFw0yMzEyMjAxOTI5MDBaFw0yMzEyMjExOTM0MDBaMCcwJQIUCIv9SIxL1N6BuSaC3S2lK0zH7WUXDTIzMTIyMTAzMzIwMFqgMDAuMB8GA1UdIwQYMBaAFHPc7sJbM+GwDfKNfXO6ttCLnK/8MAsGA1UdFAQEAgIF3TANBgkqhkiG9w0BAQsFAAOCAQEA41uhT72kk7XthiGSIOZP5D0MHh9xs8CkQ3jYlhbW9hw/rzEXhYFhtg4gaEFDdg/83Pnw5a7eU94+lJF5zohiU3WvW7QpzaB/PpPinMfJ34u2TY+d6KmLnAnxcw6kn7xoZ6FMPnC5YDP3sOOz82dS3JjeBK7fVz59VmTSUsBYH0cLW6FpBAsnI7DgKFEn070MqD89nTHlwK+hEClkznR887DHVtBP1+0pVrpH69R8XNEMflKHvWf9v6s/ceb9a9ufG6QWdp8gJ7i70VTygk8PYSSgsR0EsYy/k2sioeFt6CEwvgh8xtMBgMXM0Zo7CN/Zmmvhj+fEjjovpyA9sDabxw==</publish>
</delta>`
	deltaModel := DeltaModel{}
	err := xmlutil.UnmarshalXml(body, &deltaModel)
	fmt.Println(deltaModel, err)
	for i := range deltaModel.DeltaPublishs {
		fmt.Println("uri:", deltaModel.DeltaPublishs[i].Uri)
		fmt.Println("hash:", deltaModel.DeltaPublishs[i].Hash)
		fmt.Println("base64:", deltaModel.DeltaPublishs[i].Base64)
	}
}
func TestGetRrdpDelta(t *testing.T) {
	url := `https://rpki.luys.cloud/rrdp/b14bf9f4-2d4f-418f-8f47-819710b78724/3024/delta.xml`
	deltaModel, err := GetRrdpDelta(url)
	/*
		_, body, err := httpclient.GetHttpsVerify(url, false)
		fmt.Println(body)

		deltaModel := DeltaModel{}
		err = xmlutil.UnmarshalXml(body, &deltaModel)
	*/
	fmt.Println(deltaModel, err)

	for i := range deltaModel.DeltaPublishs {
		fmt.Println("uri:", deltaModel.DeltaPublishs[i].Uri)
		fmt.Println("hash:", deltaModel.DeltaPublishs[i].Hash)
		fmt.Println("base64:", deltaModel.DeltaPublishs[i].Base64)
	}
}

func TestGetRrdpNotification(t *testing.T) {
	notificationModel, err := GetRrdpNotification("https://rpki.idnic.net/rrdp/notify.xml")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("get notification ok")

	err = CheckRrdpNotification(&notificationModel)
	if err != nil {
		fmt.Println(err)
		return
	}
}
func TestGetRrdpSnapshot(t *testing.T) {
	url := `https://rpki-repo.registro.br/rrdp/49582cf3-79ba-4cba-a1a9-14e966177268/147100/65979eb2b415672b/snapshot.xml`
	fmt.Println("will get snapshot:", url)
	snapshotModel, err := GetRrdpSnapshot(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("get snapshot ok:", snapshotModel)
	/*
		err = CheckRrdpSnapshot(&snapshotModel, &notificationModel)
		if err != nil {
			fmt.Println(err)
			return
		}

			err = SaveRrdpSnapshotToFiles(&snapshotModel, `G:\Download\rrdp`)
			if err != nil {
				fmt.Println(err)
				return
			}

			for i, _ := range notificationModel.Deltas {

				deltaModel, err := GetRrdpDelta(notificationModel.Deltas[i].Uri)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("get delta ok")

				err = CheckRrdpDelta(&deltaModel, &notificationModel)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = SaveRrdpDeltaToFiles(&deltaModel, `G:\Download\rrdp`)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
	*/
	fmt.Println("all ok")
}

func TestGetHttpsVerifySupportRangeWithConfig(t *testing.T) {
	urls := []string{
		"https://0.sb/rrdp/notification.xml",
		"https://ca.nat.moe/rrdp/notification.xml",
		"https://ca.rg.net/rrdp/notify.xml",
		"https://chloe.sobornost.net/rpki/news.xml",
		"https://cloudie.rpki.app/rrdp/notification.xml",
		"https://dev.tw/rpki/notification.xml",
		"https://krill.accuristechnologies.ca/rrdp/notification.xml",
		"https://krill.ca-bc-01.ssmidge.xyz/rrdp/notification.xml",
		"https://krill.rayhaan.net/rrdp/notification.xml",
		"https://krill.stonham.uk/rrdp/notification.xml",
		"https://magellan.ipxo.com/rrdp/notification.xml",
		"https://pub.krill.ausra.cloud/rrdp/notification.xml",
		"https://pub.rpki.win/rrdp/notification.xml",
		"https://repo-rpki.idnic.net/rrdp/notification.xml",
		"https://repo.kagl.me/rpki/notification.xml",
		"https://rov-measurements.nlnetlabs.net/rrdp/notification.xml",
		"https://rpki-01.pdxnet.uk/rrdp/notification.xml",
		"https://rpki-publication.haruue.net/rrdp/notification.xml",
		"https://rpki-repo.registro.br/rrdp/notification.xml",
		"https://rpki-repository.nic.ad.jp/rrdp/ap/notification.xml",
		"https://rpki-rrdp.mnihyc.com/rrdp/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/08c2f264-23f9-49fb-9d43-f8b50bec9261/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/16f1ffee-7461-4674-bb05-fddefa9a02c6/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/20aa329b-fc52-4c61-bf53-09725c042942/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/2f059a21-d41b-4846-b7ae-7ea38c32fd4c/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/42582c67-dd3f-4bc5-ba60-e97e552c6e35/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/517f3ed7-58b5-4796-be37-14d62e48f056/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/54602fb0-a9d4-4f9f-b0ca-be2a139ea92b/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/602a26e5-4a9e-4e5e-89f0-ef891490d9c9/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/708aafaf-00b4-485b-854c-0b32ca30f57b/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/71e5236f-c6f1-4928-a1b9-8def09c06085/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/967a255c-d680-42d3-9ec3-ecb3f9da088c/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/a841823c-a10d-477c-bfdf-4086f0b1594c/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/b3f6b688-cff4-402f-97d5-02f6f1886b7e/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/b68a32ee-455d-483a-943d-1a5be748bfea/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/b8a1dd25-c313-4f25-ac21-bf55514d9c7d/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/bd48a1fa-3471-4ab2-8508-ad36b96813e4/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/c3cd7c24-12cb-4abc-8fd2-5e2bcbb85ae6/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/db9a372a-09bc-4a32-bfe4-8c48e5dbd219/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/dba8f01c-9669-44a3-ac6e-db2edb099b84/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/dfd7f6d3-e6e9-4987-9ae7-d052c5353898/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/e72d8db0-4728-4fc1-bdd8-471129866362/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/e7518af5-a343-428d-bf78-f982b6e60505/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/f703696e-e47b-4c20-bd93-6f80904e42d2/notification.xml",
		"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/ff9fa84e-9783-4a0b-a58d-6dc8e2433d33/notification.xml",
		"https://rpki.0i1.eu/rrdp/notification.xml",
		"https://rpki.admin.freerangecloud.com/rrdp/notification.xml",
		"https://rpki.akrn.net/rrdp/notification.xml",
		"https://rpki.apernet.io/rrdp/notification.xml",
		"https://rpki.athene-center.net/rrdp/notification.xml",
		"https://rpki.berrybyte.network/rrdp/notification.xml",
		"https://rpki.caramelfox.net/rrdp/notification.xml",
		"https://rpki.caramelfox.net/rrpdp/notification.xml",
		"https://rpki.cc/rrdp/notification.xml",
		"https://rpki.cernet.edu.cn/rrdp/notification.xml",
		"https://rpki.cnnic.cn/rrdp/notify.xml",
		"https://rpki.co/rrdp/notification.xml",
		"https://rpki.ezdomain.ru/rrdp/notification.xml",
		"https://rpki.folf.systems/rrdp/notification.xml",
		"https://rpki.komorebi.network:3030/rrdp/notification.xml",
		"https://rpki.luys.cloud/rrdp/notification.xml",
		"https://rpki.multacom.com/rrdp/notification.xml",
		"https://rpki.nap.re:3030/rrdp/notification.xml",
		"https://rpki.netiface.net/rrdp/notification.xml",
		"https://rpki.owl.net/rrdp/notification.xml",
		"https://rpki.pedjoeang.group/rrdp/notification.xml",
		"https://rpki.pudu.be/rrdp/notification.xml",
		"https://rpki.qs.nu/rrdp/notification.xml",
		"https://rpki.rand.apnic.net/rrdp/notification.xml",
		"https://rpki.roa.net/rrdp/notification.xml",
		"https://rpki.sailx.co/rrdp/notification.xml",
		"https://rpki.sn-p.io/rrdp/notification.xml",
		"https://rpki.ssmidge.xyz/rrdp/notification.xml",
		"https://rpki.telecentras.lt/rrdp/notification.xml",
		"https://rpki.tools.westconnect.ca/rrdp/notification.xml",
		"https://rpki.xindi.eu/rrdp/notification.xml",
		"https://rpki.zappiehost.com/rrdp/notification.xml",
		"https://rpki1.rpki-test.sit.fraunhofer.de/rrdp/notification.xml",
		"https://rpkica.mckay.com/rrdp/notify.xml",
		"https://rrdp-as0.apnic.net/notification.xml",
		"https://rrdp-rps.arin.net/notification.xml",
		"https://rrdp.afrinic.net/notification.xml",
		"https://rrdp.apnic.net/notification.xml",
		"https://rrdp.arin.net/notification.xml",
		"https://rrdp.krill.cloud/notification.xml",
		"https://rrdp.lacnic.net/rrdp/notification.xml",
		"https://rrdp.lacnic.net/rrdpas0/notification.xml",
		"https://rrdp.paas.rpki.ripe.net/notification.xml",
		"https://rrdp.ripe.net/notification.xml",
		"https://rrdp.roa.tohunet.com/rrdp/notification.xml",
		"https://rrdp.rp.ki/notification.xml",
		"https://rrdp.rpki.co/rrdp/notification.xml",
		"https://rrdp.rpki.tianhai.link/rrdp/notification.xml",
		"https://rrdp.sub.apnic.net/notification.xml",
		"https://rrdp.taaa.eu/rrdp/notification.xml",
		"https://rrdp.twnic.tw/rrdp/notify.xml",
		"https://rsync.rpki.co/rrdp/notification.xml",
		"https://sakuya.nat.moe/rrdp/notification.xml",
		"https://x-0100000000000011.p.u9sv.com/notification.xml",
		"https://x-8011.p.u9sv.com/notification.xml",
	}
	h := httpclient.NewHttpClientConfigWithParam(5, 3, "all")
	supportUrls := make([]string, 0, len(urls))
	for _, url := range urls {
		n, err := GetRrdpNotificationWithConfig(url, h)
		if err != nil {
			fmt.Println("fail, url:", url, err)
			continue
		}
		snapshotUrl := n.Snapshot.Uri
		_, supportRange, contentLength, err := httpclient.GetHttpsVerifySupportRangeWithConfig(snapshotUrl, true, h)
		if err != nil {
			fmt.Println("fail, url:", url, snapshotUrl, err)
			continue
		}
		if supportRange && contentLength > 10000000 {
			fmt.Println("support: ", url, snapshotUrl, contentLength)
			supportUrls = append(supportUrls, url)
		}
	}
	fmt.Println(jsonutil.MarshalJson(supportUrls))
	/*
		[
			"https://repo-rpki.idnic.net/rrdp/notification.xml",
			"https://rpki-repo.registro.br/rrdp/notification.xml",
			"https://rpki-repository.nic.ad.jp/rrdp/ap/notification.xml",
			"https://rpki-rrdp.us-east-2.amazonaws.com/rrdp/20aa329b-fc52-4c61-bf53-09725c042942/notification.xml",
			"https://rrdp.lacnic.net/rrdp/notification.xml",
			"https://rrdp.twnic.tw/rrdp/notify.xml"
		]

	*/
}
