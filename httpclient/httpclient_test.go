package httpclient

import (
	"fmt"
	"testing"

	"github.com/cpusoft/goutil/jsonutil"
)

func TestGetHttpsVerifyWithConfig(t *testing.T) {
	url := `https://rrdp.ripe.net/notification.xml`
	resp, body, err := GetHttpsWithConfig(url, NewHttpClientConfigWithParam(5, 3, "all", true))
	fmt.Println("res:", resp)
	fmt.Println("body:", body)
	fmt.Println("err:", err)
}
func TestGetHttpsResponseVerifyWithConfig(t *testing.T) {
	url := `https://rrdp.ripe.net/notification.xml`
	url = `https://rrdp.ripe.net/172322cf-c642-4e6f-806c-bd2375d8001a/62034/snapshot-ed067615f1f801318d6233f9dd89aa204e250cdc30ced98918ceadbfcbc3d173.xml`
	url = `https://rpki-repo.registro.br/rrdp/49582cf3-79ba-4cba-a1a9-14e966177268/137263/65979eb2b415672b/snapshot.xml`
	resp, _ := GetHttpsResponseWithConfig(url, NewHttpClientConfigWithParam(5, 3, "all", true))
	fmt.Println("res:", jsonutil.MarshalJson(resp.Header))
	ar := resp.Header.Get("Accept-Ranges")
	len := resp.Header.Get("Content-Length")
	fmt.Println("ar:", ar, "  len:", len)
}

func TestGetHttpsRangeVerifyWithConfig(t *testing.T) {
	// support range
	// support gzip

	url := `https://rpki-repo.registro.br/rrdp/49582cf3-79ba-4cba-a1a9-14e966177268/147100/65979eb2b415672b/snapshot.xml`
	_, supportRange, contentLength, err := GetHttpsSupportRangeWithConfig(url, NewHttpClientConfigWithParam(5, 3, "all", true))
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	if !supportRange {
		fmt.Println("not suppport:")
		return
	}
	fmt.Println("contentLength:", contentLength)
	resp, body, err := GetHttpsRangeWithConfig(url, contentLength,
		10000000, NewHttpClientConfigWithParam(5, 3, "all", true))
	fmt.Println("res:", jsonutil.MarshalJson(resp), " status:", GetStatusCode(resp))
	fmt.Println("body:", len(body))
	fmt.Println("err:", err)
}

func TestPostHttp(t *testing.T) {
	urlStr := `https://202.173.14.105:8070/parsevalidate/parsefilesimple`
	postJson := ``
	res, body, err := Post(urlStr, postJson, false)
	fmt.Println("res:", res)
	fmt.Println("body:", body)
	fmt.Println("err:", err)
}

func TestGetHttp(t *testing.T) {
	/*
		//往rp发送请求
		resp, body, err := GetHttp("http://localhost:8080/hello")

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
	*/
}

func TestGetHttps(t *testing.T) {
	/*
		//往rp发送请求
		resp, body, err := GetHttps("https://localhost:8081/hello")

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
	*/
}

// only in linux
func TestGetHttpsRrdp(t *testing.T) {

	//往rp发送请求
	//`https://rrdp.afrinic.net/notification.xml` //https://rpki.august.tw/rrdp/notification.xml`
	url := `https://rrdp-as0.apnic.net/e197f36e-b1c0-46f8-a2f6-ffc00cf83c38/44292/snapshot.xml`
	url = `https://rrdp.ripe.net/notification.xml`
	url = `https://rrdp.arin.net/4a394319-7460-4141-a416-1addb69284ff/67365/snapshot.xml`
	//SetTimeout(30)
	//defer ResetTimeout()
	resutl, err := GetByCurlWithConfig(url, NewHttpClientConfigWithParam(5, 3, "ipv4", true))

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp:", resutl)

	resutl, err = GetByCurlWithConfig(url, NewHttpClientConfigWithParam(5, 3, "all", true))

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("iptype is all, resp:", resutl)

}

func TestPostFile(t *testing.T) {
	/*
		resp, body, err := PostFile("http", "localhost", 8080, "/parse/start",
			`G:/Download/cert/cache/trustanchors/rpki.apnic.net/repository/apnic-rpki-root-iana-origin.cer`, `file`)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp:", resp)
		fmt.Println("body:", body)
	*/
}

func TestHttpsPostFile(t *testing.T) {

	resp, body, err := PostFileHttpWithConfig("http://202.173.14.105:8070/parsevalidate/parsefilesimple",
		`G:\Download\cert\oWhEB7GUTj5ZqlXo7X2VbNrJ9xw.cer`, `file`, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp:", resp)
	fmt.Println("body:", body)

}

func TestHttpUnMarshalJson(t *testing.T) {
	/*
		url := `http://202.173.14.103:58085/allReset`
		var httpResponse HttpResponse
		err := PostAndUnmarshalStruct(url, "", false, &httpResponse)
		fmt.Println(httpResponse, err)

		resp, body, err := Post(url, "", false)
		resp.Body.Close()
		var httpResponse1 HttpResponse
		err = jsonutil.UnmarshalJson(body, &httpResponse1)
		fmt.Println(httpResponse1, err)
	*/
}

type HttpResponse struct {
	Result string      `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

// go test -v -test.run TestDownloadUrlFile -timeout 50m
func TestDownloadUrlFile(t *testing.T) {
	urlFile := `https://data.ris.ripe.net/rrc00/2020.01/bview.20200105.0000.gz`
	localFile := `./1.gz`
	n, err := DownloadUrlFile(urlFile, localFile)
	fmt.Println(n, err)

}
