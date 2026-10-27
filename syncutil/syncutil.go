package syncutil

import (
	"errors"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/httpclient2"
	"github.com/cpusoft/goutil/rrdputil"
	"github.com/cpusoft/goutil/rsyncutil"
)

func TestRepoConnect(repoUrl string, timeoutMins uint64) (state string, DurationTime time.Duration, err error) {
	belogs.Debug("TestRepoConnect(): repoUrl:", repoUrl, " timeoutMins:", timeoutMins)

	if strings.HasPrefix(repoUrl, "https://") {
		belogs.Debug("TestRepoConnect(): rrdp repoUrl:", repoUrl, " timeoutMins:", timeoutMins)
		start := time.Now()
		cfg := httpclient2.NewHttpClientConfig(
			httpclient2.WithAllTimeoutSeconds(30),
			httpclient2.WithTlsHandshakeTimeoutSeconds(5),
			httpclient2.WithConnectionTimeoutSeconds(3),
			httpclient2.WithRetryCount(2),
			httpclient2.WithIpType("ipv4"),
			httpclient2.WithVerifyHttps(true),
		)
		err := rrdputil.RrdpNotificationTestConnectWithConfig(repoUrl, cfg)
		if err != nil {
			belogs.Error("TestRepoConnect(): RrdpNotificationTestConnectWithConfig fail, repoUrl:", repoUrl, err)
			return "invalid", time.Since(start), err
		} else {
			belogs.Debug("TestRepoConnect(): RrdpNotificationTestConnectWithConfig ok,, repoUrl:", repoUrl)
			return "valid", time.Since(start), nil
		}

	} else if strings.HasPrefix(repoUrl, "rsync://") {
		belogs.Debug("TestRepoConnect(): rsync repoUrl:", repoUrl)
		start := time.Now()
		err := rsyncutil.RsyncTestConnect(repoUrl)
		if err != nil {
			belogs.Error("TestRepoConnect(): RsyncTestConnect fail, repoUrl:", repoUrl, err)
			return "invalid", time.Since(start), err
		} else {
			belogs.Debug("TestRepoConnect(): RsyncTestConnect ok,, repoUrl:", repoUrl)
			return "valid", time.Since(start), nil
		}
	} else {
		start := time.Now()
		belogs.Error("TestRepoConnect(): protocol is wrong fail, repoUrl:", repoUrl)
		return "invalid", time.Since(start), errors.New("repoUrl's protocol is wrong")
	}

}
