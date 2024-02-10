package syncutil

import (
	"errors"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/rrdputil"
	"github.com/cpusoft/goutil/rsyncutil"
)

func TestConnnect(repoUrl string, timeoutMins uint64) (state string, DurationTime time.Duration, err error) {
	belogs.Debug("TestConnnect(): repoUrl:", repoUrl, " timeoutMins:", timeoutMins)

	if strings.HasPrefix(repoUrl, "https://") {
		belogs.Debug("TestConnnect(): rrdp repoUrl:", repoUrl, " timeoutMins:", timeoutMins)
		start := time.Now()
		err := rrdputil.RrdpNotificationTestConnectWithConfig(repoUrl,
			httpclient.NewHttpClientConfigWithParam(uint64(timeoutMins), 1, "all"))
		if err != nil {
			belogs.Error("TestConnnect(): RrdpNotificationTestConnectWithConfig fail, repoUrl:", repoUrl, err)
			return "invalid", time.Since(start), err
		} else {
			belogs.Debug("TestConnnect(): RrdpNotificationTestConnectWithConfig ok,, repoUrl:", repoUrl)
			return "valid", time.Since(start), nil
		}

	} else if strings.HasPrefix(repoUrl, "rsync://") {
		belogs.Debug("TestConnnect(): rsync repoUrl:", repoUrl)
		start := time.Now()
		err := rsyncutil.RsyncTestConnect(repoUrl)
		if err != nil {
			belogs.Error("TestConnnect(): RsyncTestConnect fail, repoUrl:", repoUrl, err)
			return "invalid", time.Since(start), err
		} else {
			belogs.Debug("TestConnnect(): RsyncTestConnect ok,, repoUrl:", repoUrl)
			return "valid", time.Since(start), nil
		}
	} else {
		start := time.Now()
		belogs.Error("TestConnnect(): protocol is wrong fail, repoUrl:", repoUrl)
		return "invalid", time.Since(start), errors.New("repoUrl's protocol is wrong")
	}

}
