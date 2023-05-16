package rrdputil

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/netutil"
	"github.com/cpusoft/goutil/xmlutil"
)

func GetRrdpNotification(notificationUrl string) (notificationModel NotificationModel, err error) {
	start := time.Now()
	// get notification.xml
	// "https://rrdp.apnic.net/notification.xml"
	belogs.Info("GetRrdpNotification(): will notificationUrl:", notificationUrl)
	notificationModel, err = getRrdpNotificationImpl(notificationUrl)
	if err != nil {
		belogs.Error("GetRrdpNotification():getRrdpNotificationImpl fail:", notificationUrl, err)
		return notificationModel, err
	}

	// will sort deltas from bigger to smaller
	sort.Sort(NotificationDeltasSort(notificationModel.Deltas))
	belogs.Debug("GetRrdpNotification(): after sort, len(notificationModel.Deltas):", len(notificationModel.Deltas))

	// get maxserial and minserial, and set map[serial]serial
	notificationModel.MapSerialDeltas = make(map[uint64]uint64, len(notificationModel.Deltas)+10)
	for i := range notificationModel.Deltas {
		notificationModel.MapSerialDeltas[notificationModel.Deltas[i].Serial] = notificationModel.Deltas[i].Serial
		serial := notificationModel.Deltas[i].Serial
		if serial > notificationModel.MaxSerial {
			notificationModel.MaxSerial = serial
		}
		if serial < notificationModel.MinSerial {
			notificationModel.MinSerial = serial
		}
	}
	belogs.Info("GetRrdpNotification(): notificationUrl ok:", notificationUrl, "  time(s):", time.Now().Sub(start).Seconds())
	return notificationModel, nil
}

func getRrdpNotificationImpl(notificationUrl string) (notificationModel NotificationModel, err error) {

	belogs.Debug("getRrdpNotificationImpl(): notificationUrl:", notificationUrl)
	start := time.Now()
	notificationUrl = strings.TrimSpace(notificationUrl)
	resp, body, err := httpclient.GetHttpsVerify(notificationUrl, true)
	if err == nil {
		defer resp.Body.Close()
		belogs.Debug("getRrdpNotificationImpl(): GetHttpsVerify notificationUrl:", notificationUrl,
			"   ipAddrs:", netutil.LookupIpByUrl(notificationUrl), "   resp.Status:", resp.Status,
			"   len(body):", len(body),
			"   time(s):", time.Now().Sub(start).Seconds())

		if resp.StatusCode != http.StatusOK {
			belogs.Error("getRrdpNotificationImpl(): GetHttpsVerify notificationUrl, is not StatusOK:", notificationUrl,
				"   resp.Status:", resp.Status, "    body:", body)
			return notificationModel, errors.New("http status code of " + notificationUrl + " is " + resp.Status)
		}

	} else {
		belogs.Error("getRrdpNotificationImpl(): GetHttpsVerify notificationUrl fail, will use curl again:", notificationUrl, "   resp:",
			resp, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)

		// then try using curl
		start = time.Now()
		body, err = httpclient.GetByCurl(notificationUrl)
		if err != nil {
			belogs.Error("getRrdpNotificationImpl(): GetByCurl notificationUrl fail:", notificationUrl,
				"   ipAddrs:", netutil.LookupIpByUrl(notificationUrl), "   resp:", resp,
				"   len(body):", len(body), "       body:", body, "  time(s):", time.Now().Sub(start).Seconds(), err)
			return notificationModel, errors.New("http error of " + notificationUrl + " is " + err.Error())
		}
		belogs.Debug("getRrdpNotificationImpl(): GetByCurl deltaUrl ok", notificationUrl, "    len(body):", len(body),
			"  time(s):", time.Now().Sub(start).Seconds())
	}
	// check if body is xml file
	if !strings.Contains(body, `<notification`) {
		belogs.Error("getRrdpNotificationImpl(): body is not xml file:", notificationUrl, "   resp:",
			resp, "    len(body):", len(body), "       body:", body, "  time(s):", time.Now().Sub(start).Seconds(), err)
		return notificationModel, errors.New("body of " + notificationUrl + " is not xml")
	}

	// unmarshal xml
	err = xmlutil.UnmarshalXml(body, &notificationModel)
	if err != nil {
		belogs.Error("getRrdpNotificationImpl(): UnmarshalXml fail: ", notificationUrl, "        body:", body, err)
		return notificationModel, errors.New("response of " + notificationUrl + " is not a legal rrdp file")
	}
	return notificationModel, nil
}

func RrdpNotificationTestConnect(notificationUrl string) (err error) {
	start := time.Now()
	belogs.Debug("RrdpNotificationTestConnect(): notificationUrl:", notificationUrl)

	// test http connect
	resp, body, err := httpclient.GetHttpsVerify(notificationUrl, true)
	if err != nil {
		belogs.Error("RrdpNotificationTestConnect(): GetHttpsVerify fail, notificationUrl:", notificationUrl, err, "  time(s):", time.Since(start))
		return errors.New("http error of " + notificationUrl + " is " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		belogs.Error("RrdpNotificationTestConnect(): GetHttpsVerify notificationUrl, is not StatusOK:", notificationUrl,
			"   resp.Status:", resp.Status, "    body:", body)
		return errors.New("http status code of " + notificationUrl + " is " + resp.Status)
	}
	belogs.Debug("RrdpNotificationTestConnect(): GetHttpsVerify ok, notificationUrl:", notificationUrl, "  time(s):", time.Since(start))

	// test is legal
	var notificationModel NotificationModel
	err = xmlutil.UnmarshalXml(body, &notificationModel)
	if err != nil {
		belogs.Error("RrdpNotificationTestConnect(): UnmarshalXml to get notificationModel fail, notificationUrl:", notificationUrl, "     body:", body, err)
		return errors.New("response of " + notificationUrl + " is not a legal rrdp file")
	}
	belogs.Info("RrdpNotificationTestConnect(): get notificationModel ok, notificationUrl:", notificationUrl,
		"   time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func CheckRrdpNotification(notificationModel *NotificationModel) (err error) {
	if notificationModel.Version != "1" {
		belogs.Error("CheckRrdpNotification():  notificationModel.Version != 1")
		return errors.New("notification version is error, version is not 1, it is " + notificationModel.Version)
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
