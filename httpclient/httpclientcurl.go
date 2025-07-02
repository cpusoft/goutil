package httpclient

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/netutil"
)

/*
func GetByCurl(url string) (result string, err error) {
	return GetByCurlWithConfig(url, nil)
}
*/
// get by Curl
func GetByCurlWithConfig(url string, httpClientConfig *HttpClientConfig) (result string, err error) {
	belogs.Debug("GetByCurlWithConfig(): url:", url, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	url = strings.TrimSpace(url)
	if len(url) == 0 {
		return "", errors.New("url is emtpy")
	}
	if httpClientConfig == nil {
		httpClientConfig = NewHttpClientConfig()
	}
	// mins --> seconds
	timeout := convert.ToString(httpClientConfig.TimeoutMins * 60)
	retryCount := convert.ToString(httpClientConfig.RetryCount)
	//	tmpFile := os.TempDir() + string(os.PathSeparator) + uuidutil.GetUuid()
	tmpFile, err := os.CreateTemp("", "_tmp_") // temp file
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): TempFile fail", err)
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	ipType := ""
	if httpClientConfig.IpType == "ipv4" {
		ipType = "-4"
	} else if httpClientConfig.IpType == "ipv6" {
		ipType = "-6"
	}

	belogs.Info("GetByCurlWithConfig():will curl, url:", url, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig),
		"  httpClientConfig.TimeoutMins(m):", int64(httpClientConfig.TimeoutMins), "  timeout as seconds:", timeout,
		"  retryCount:", retryCount, "  ipType:", ipType, "   tmpFile:", tmpFile.Name())

	// -s: slient mode  --no use
	// -4: ipv4  --no use
	// --connect-timeout: SECONDS  Maximum time allowed for connection
	// --ignore-content-length: Ignore the Content-Length header  --no use
	// --retry:
	// -o : output file
	// --limit-rate:  100k  --no use
	// --keepalive-time: <seconds> Interval time for keepalive probes
	// -m: --max-time SECONDS  Maximum time allowed for the transfer
	// -v, --verbose       Make the operation more talkative
	// -k, --insecure      Allow connections to SSL sites without certs (H)
	/*
		cmd := exec.Command("curl", "-4",  "-o", tmpFile, url)
	*/
	// minute-->second
	var args []string
	args = append(args, "--connect-timeout", timeout, "--keepalive-time", timeout, "-m", timeout)

	// curl high version warning:  should  filter ipType null
	if ipType != "" {
		args = append(args, ipType)
	}
	if !httpClientConfig.VerifyHttps {
		args = append(args, "--insecure")
	}
	args = append(args, "--retry", retryCount, "--compressed", "-o", tmpFile.Name(), url)

	start := time.Now()
	cmd := exec.Command("curl", args...)
	output, err := cmd.CombinedOutput()
	outputStr := GetOutputStr(output)
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): exec.Command fail, curl:", url, "  ipAddrs:", netutil.LookupIpByUrl(url),
			"  tmpFile:", tmpFile.Name(), "  timeout:", timeout, "  retryCount:", retryCount, "  ipType:", ipType,
			"  len(output):", len(output), "  outputStr:", outputStr, "  time(s):", time.Since(start), "   err:", err)
		return "", errors.New("Fail to get by curl. url is " + url + ". " + err.Error())
	}
	belogs.Debug("GetByCurlWithConfig(): curl ok, url:", url, "   tmpFile:", tmpFile.Name(), "  timeout:", timeout,
		"  len(output):", len(output), "  outputStr:", outputStr, "  time(s):", time.Since(start))

	b, err := fileutil.ReadFileToBytes(tmpFile.Name())
	if err != nil {
		belogs.Error("GetByCurlWithConfig(): ReadFileToBytes fail, url", url, "   tmpFile:", tmpFile.Name(), "   err: ", err, "   output: "+string(output))
		return "", errors.New("Fail to get by curl. Error is `" + err.Error() + "`. Output  is `" + string(output) + "`")
	}
	belogs.Debug("GetByCurlWithConfig(): ReadFileToBytes ok, url:", url, "   tmpFile:", tmpFile.Name(), "  len(b):", len(b), "  time(s):", time.Since(start))
	return string(b), nil
}
