package rrdputil

import (
	"fmt"
	"testing"

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
	url := `https://rrdp.arin.net/8fe05c2e-047d-49e7-8398-cd4250a572b1/7677/snapshot.xml`
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
