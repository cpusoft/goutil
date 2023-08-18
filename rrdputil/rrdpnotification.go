package rrdputil

import (
	"errors"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/netutil"
	"github.com/cpusoft/goutil/xmlutil"
)

func GetRrdpNotification(notificationUrl string) (notificationModel NotificationModel, err error) {
	belogs.Debug("GetRrdpNotification(): will notificationUrl:", notificationUrl)
	return GetRrdpNotificationWithConfig(notificationUrl, nil)
}

func GetRrdpNotificationWithConfig(notificationUrl string, httpClientConfig *httpclient.HttpClientConfig) (notificationModel NotificationModel, err error) {
	start := time.Now()
	// get notification.xml
	// "https://rrdp.apnic.net/notification.xml"
	belogs.Info("GetRrdpNotificationWithConfig(): will notificationUrl:", notificationUrl, "   httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	notificationModel, err = getRrdpNotificationImplWithConfig(notificationUrl, httpClientConfig)
	if err != nil {
		belogs.Error("GetRrdpNotificationWithConfig():getRrdpNotificationImplWithConfig fail:", notificationUrl, err)
		return notificationModel, err
	}

	// will sort deltas from bigger to smaller
	sort.Sort(NotificationDeltasSort(notificationModel.Deltas))
	belogs.Debug("GetRrdpNotificationWithConfig(): after sort, len(notificationModel.Deltas):", len(notificationModel.Deltas))

	// get maxserial and minserial, and set map[serial]serial
	notificationModel.MapSerialDeltas = make(map[uint64]uint64, len(notificationModel.Deltas)+10)
	min := uint64(math.MaxUint64)
	max := uint64(0)
	for i := range notificationModel.Deltas {
		notificationModel.MapSerialDeltas[notificationModel.Deltas[i].Serial] = notificationModel.Deltas[i].Serial
		serial := notificationModel.Deltas[i].Serial
		if serial > max {
			max = serial
		}
		if serial < min {
			min = serial
		}
	}
	notificationModel.MaxSerial = max
	notificationModel.MinSerial = min
	belogs.Info("GetRrdpNotificationWithConfig(): notificationUrl ok:", notificationUrl, " notificationModel.MaxSerial:", notificationModel.MaxSerial,
		"   notificationModel.MinSerial:", notificationModel.MinSerial, "  time(s):", time.Since(start))
	return notificationModel, nil
}

func getRrdpNotificationImplWithConfig(notificationUrl string, httpClientConfig *httpclient.HttpClientConfig) (notificationModel NotificationModel, err error) {

	belogs.Debug("getRrdpNotificationImplWithConfig(): notificationUrl:", notificationUrl, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	start := time.Now()
	notificationUrl = strings.TrimSpace(notificationUrl)
	resp, body, err := httpclient.GetHttpsVerifyWithConfig(notificationUrl, true, httpClientConfig)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			belogs.Error("getRrdpNotificationImplWithConfig(): GetHttpsVerifyWithConfig notificationUrl, is not StatusOK:", notificationUrl,
				"   resp.Status:", resp.Status, "    body:", body)
			return notificationModel, errors.New("http status code of " + notificationUrl + " is " + resp.Status)
		} else {
			belogs.Debug("getRrdpNotificationImplWithConfig(): GetHttpsVerifyWithConfig notificationUrl:", notificationUrl,
				"   ipAddrs:", netutil.LookupIpByUrl(notificationUrl), "   resp.Status:", resp.Status,
				"   len(body):", len(body),
				"   time(s):", time.Since(start))
		}

	} else {
		belogs.Debug("getRrdpNotificationImplWithConfig(): GetHttpsVerifyWithConfig notificationUrl fail, will use curl again:", notificationUrl, "   resp:",
			resp, "    len(body):", len(body), "  time(s):", time.Since(start), err)

		// then try using curl
		start = time.Now()
		body, err = httpclient.GetByCurlWithConfig(notificationUrl, httpClientConfig)
		if err != nil {
			belogs.Debug("getRrdpNotificationImplWithConfig(): GetByCurlWithConfig notificationUrl, iptype is ipv4, fail:", notificationUrl,
				"   ipAddrs:", netutil.LookupIpByUrl(notificationUrl), "   resp:", resp,
				"   len(body):", len(body), "       body:", body, "  time(s):", time.Since(start), err)

			// then try again using curl, using all
			start = time.Now()
			httpClientConfig.IpType = "all"
			body, err = httpclient.GetByCurlWithConfig(notificationUrl, httpClientConfig)
			if err != nil {
				belogs.Error("getRrdpNotificationImplWithConfig(): GetByCurlWithConfig notificationUrl, iptype is all, fail:", notificationUrl,
					"   ipAddrs:", netutil.LookupIpByUrl(notificationUrl), "   resp:", resp,
					"   len(body):", len(body), "  time(s):", time.Since(start), err)
				return notificationModel, errors.New("http error of " + notificationUrl + " is " + err.Error())
			}
			belogs.Debug("getRrdpNotificationImplWithConfig(): GetByCurlWithConfig notificationUrl, iptype is all, ok", notificationUrl, "    len(body):", len(body),
				"  time(s):", time.Since(start))
		} else {
			belogs.Debug("getRrdpNotificationImplWithConfig(): GetByCurlWithConfig notificationUrl, iptype is ipv4, ok", notificationUrl, "    len(body):", len(body),
				"  time(s):", time.Since(start))
		}
	}
	// check if body is xml file
	if !strings.Contains(body, `<notification`) {
		belogs.Error("getRrdpNotificationImplWithConfig(): body is not xml file:", notificationUrl, "   resp:",
			resp, "    len(body):", len(body), "       body:", body, "  time(s):", time.Since(start), err)
		return notificationModel, errors.New("body of " + notificationUrl + " is not xml")
	}

	// unmarshal xml
	err = xmlutil.UnmarshalXml(body, &notificationModel)
	if err != nil {
		belogs.Error("getRrdpNotificationImplWithConfig(): UnmarshalXml fail: ", notificationUrl, "        body:", body, err)
		return notificationModel, errors.New("response of " + notificationUrl + " is not a legal rrdp file")
	}
	notificationModel.NotificationUrl = notificationUrl
	belogs.Info("getRrdpNotificationImplWithConfig(): get from notificationUrl ok", notificationUrl, "  time(s):", time.Since(start))
	return notificationModel, nil
}

func RrdpNotificationTestConnectWithConfig(notificationUrl string, httpClientConfig *httpclient.HttpClientConfig) (err error) {
	start := time.Now()
	belogs.Debug("RrdpNotificationTestConnectWithConfig(): notificationUrl:", notificationUrl, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))

	// test http connect
	resp, body, err := httpclient.GetHttpsVerifyWithConfig(notificationUrl, true, httpClientConfig)
	if err != nil {
		belogs.Error("RrdpNotificationTestConnectWithConfig(): GetHttpsVerify fail, notificationUrl:", notificationUrl, err, "  time(s):", time.Since(start))
		return errors.New("http error of " + notificationUrl + " is " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		belogs.Error("RrdpNotificationTestConnectWithConfig(): GetHttpsVerify notificationUrl, is not StatusOK:", notificationUrl,
			"   resp.Status:", resp.Status, "    body:", body, "   time(s):", time.Since(start))
		return errors.New("http status code of " + notificationUrl + " is " + resp.Status)
	}
	belogs.Debug("RrdpNotificationTestConnectWithConfig(): GetHttpsVerify ok, notificationUrl:", notificationUrl,
		"  time(s):", time.Since(start))

	// test is legal
	var notificationModel NotificationModel
	err = xmlutil.UnmarshalXml(body, &notificationModel)
	if err != nil {
		belogs.Error("RrdpNotificationTestConnectWithConfig(): UnmarshalXml to get notificationModel fail, notificationUrl:", notificationUrl,
			"     body:", body, err, "   time(s):", time.Since(start))
		return errors.New("response of " + notificationUrl + " is not a legal rrdp file")
	}
	belogs.Info("RrdpNotificationTestConnectWithConfig(): get notificationModel ok, notificationUrl:", notificationUrl,
		"   time(s):", time.Since(start))
	return nil
}

func CheckRrdpNotification(notificationModel *NotificationModel) (err error) {
	if notificationModel.Version != "1" {
		belogs.Error("CheckRrdpNotification():  notificationModel.Version != 1, get notification.xml fail, NotificationUrl is ", notificationModel.NotificationUrl)
		return errors.New("get notification.xml fail, NotificationUrl is " + notificationModel.NotificationUrl)
	}
	if len(notificationModel.SessionId) == 0 {
		belogs.Error("CheckRrdpNotification(): len(notificationModel.SessionId) == 0")
		return errors.New("notification session_id is error, session_id is empty ")
	}
	if notificationModel.Serial == 0 {
		belogs.Error("CheckRrdpNotification(): len(notificationModel.Serial) == 0")
		return errors.New("notification serial is error, serial is empty ")
	}
	if len(notificationModel.MapSerialDeltas) > 0 {
		if _, ok := notificationModel.MapSerialDeltas[notificationModel.Serial]; !ok {
			belogs.Error("CheckRrdpNotification(): notification has not such serial in deltas:", notificationModel.Serial)
			return errors.New("notification has not such serial in deltas")
		}
	}
	return nil
}
