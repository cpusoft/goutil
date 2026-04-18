package rrdputil

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xmlutil"
)

func TestSnapshot(t *testing.T) {
	// 解析 800MB XML
	xml := `
<snapshot session_id="c90ce58f-7675-4d72-9c9a-d615359ae3ba" serial="1263" version="1"
    xmlns="http://www.ripe.net/rpki/rrdp">
    <publish uri="rsync://rpki-rsync.us-east-2.amazonaws.com/volume/b3f6b688-cff4-402f-97d5-02f6f1886b7e/2592ade6-505a-4e24-928b-9a1e71b309ae.roa">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</publish>
    <publish uri="rsync://rpki-rsync.us-east-2.amazonaws.com/volume/b3f6b688-cff4-402f-97d5-02f6f1886b7e/a8ff9337-21d8-4d5e-b988-d1a983d73aea.roa">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</publish>
    <publish uri="rsync://rpki-rsync.us-east-2.amazonaws.com/volume/b3f6b688-cff4-402f-97d5-02f6f1886b7e/c518c762-ddbf-4352-b9b9-1484318adaa8.roa">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</publish>
    <publish uri="rsync://rpki-rsync.us-east-2.amazonaws.com/volume/b3f6b688-cff4-402f-97d5-02f6f1886b7e/5d7wm5lPjPYLrYyQKcAGV3ULImc.crl">MIIBuDCBoQIBATANBgkqhkiG9w0BAQsFADA9MTswOQYDVQQDEzI2ZjliOTg1YjBmZTVkZWYwOWI5OTRmOGNmNjBiYWQ4YzkwMjljMDA2NTc3NTBiMjI2NxcNMjYwNDE4MDAzODA0WhcNMjYwNDIyMDAzODA0WqAwMC4wHwYDVR0jBBgwFoAUbcpl0HFNfvJWkLwJE9NU26yJKl4wCwYDVR0UBAQCAgTsMA0GCSqGSIb3DQEBCwUAA4IBAQBYiLtxsjWhiYxD0mFeVz2TDlkHJiEckmKh8E3MpoWq+ln2h7YN+qzNUcM5QDmYJF96vp6gfwK5PmxtB32e9WP466zHeGkjE9ZsgdEfOVESkAjv8jkK2nOn2kPVmxkT2Uv0EnlAxANAgdpfN7LHK4k/I4UIY10k/B0T+QNkvamUdWOAeMeK3PUEQtBuwkL53vZyOlcoGFLPMmh+mRWbI29NnPKtTjcsa8F3IRwZnEFNG6spgyr9RFaK8pe3Rhab50kiXBrWQbniUkDJd+7dmuxpK8xkrgc9xNwJ7SGsNaC1nIyqf1Fz9pZokohTAcwLTU2P96cyBVSoHU77clDyd+Zb</publish>
    <publish uri="rsync://rpki-rsync.us-east-2.amazonaws.com/volume/b3f6b688-cff4-402f-97d5-02f6f1886b7e/5d7wm5lPjPYLrYyQKcAGV3ULImc.mft">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</publish>
</snapshot>
	`
	snapshotModel, err := parseXmlToSnapshotModel([]byte(xml))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(jsonutil.MarshalJson(snapshotModel))
}

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

	fmt.Println(deltaModel, err)

	for i := range deltaModel.DeltaPublishs {
		fmt.Println("uri:", deltaModel.DeltaPublishs[i].Uri)
		fmt.Println("hash:", deltaModel.DeltaPublishs[i].Hash)
		fmt.Println("base64:", deltaModel.DeltaPublishs[i].Base64)
	}
}

func TestGetRrdpNotification(t *testing.T) {
	notificationModel, err := GetRrdpNotificationWithConfig("https://rpki.idnic.net/rrdp/notify.xml", httpclient.NewHttpClientConfig())
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
	snapshotModel, err := GetRrdpSnapshotWithConfig(url, httpclient.NewHttpClientConfig())
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
	h := httpclient.NewHttpClientConfigWithParam(5, 3, "all", true)
	supportUrls := make([]string, 0, len(urls))
	for _, url := range urls {
		n, err := GetRrdpNotificationWithConfig(url, h)
		if err != nil {
			fmt.Println("fail, url:", url, err)
			continue
		}
		snapshotUrl := n.Snapshot.Uri
		_, supportRange, contentLength, err := httpclient.GetHttpsSupportRangeWithConfig(snapshotUrl, h)
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
