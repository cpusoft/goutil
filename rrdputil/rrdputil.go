package rrdputil

import (
	"errors"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	convert "github.com/cpusoft/goutil/convert"
	hashutil "github.com/cpusoft/goutil/hashutil"
	httpclient "github.com/cpusoft/goutil/httpclient"
	xmlutil "github.com/cpusoft/goutil/xmlutil"
)

func GetRrdpNotification(notificationUrl string) (notificationModel NotificationModel, err error) {

	// 往rp发送请求
	// "https://rrdp.apnic.net/notification.xml"
	belogs.Debug("GetRrdpNotification(): notificationUrl:", notificationUrl)
	resp, body, err := httpclient.GetHttps(notificationUrl)
	if err != nil {
		belogs.Error("GetRrdpNotification(): notificationUrl fail, ", notificationUrl, err)
		return notificationModel, err
	}
	belogs.Debug("GetRrdpNotification(): resp.Status:", resp.Status, "    len(body):", len(body))

	err = xmlutil.UnmarshalXml(body, &notificationModel)
	if err != nil {
		belogs.Error("GetRrdpNotification(): UnmarshalXml fail, ", notificationUrl, err)
		return notificationModel, err
	}

	notificationModel.MapSerialDeltas = make(map[string]NotificationDelta, len(notificationModel.Deltas)+10)
	for i, _ := range notificationModel.Deltas {
		notificationModel.MapSerialDeltas[notificationModel.Deltas[i].Serial] = notificationModel.Deltas[i]
		serial := convert.Bytes2Uint64([]byte(notificationModel.Deltas[i].Serial))
		if serial > notificationModel.MaxSerail {
			notificationModel.MaxSerail = serial
		}
		if serial < notificationModel.MinSeail {
			notificationModel.MinSeail = serial
		}
	}
	//clear notificationModel.Deltas
	notificationModel.Deltas = make([]NotificationDelta, 0)
	return notificationModel, nil
}
func CheckRrdpNotification(notificationModel *NotificationModel) (err error) {
	if notificationModel.Version != "1" {
		belogs.Error("CheckRrdpNotification():  notificationModel.Version != 1")
		return errors.New("notification version is error, version is not 1, it is " + notificationModel.Version)
	}
	if len(notificationModel.Session_id) == 0 {
		belogs.Error("CheckRrdpNotification(): len(notificationModel.Session_id) == 0")
		return errors.New("notification session_id is error, session_id is empty ")
	}
	if len(notificationModel.Serial) == 0 {
		belogs.Error("CheckRrdpNotification(): len(notificationModel.Serial) == 0")
		return errors.New("notification serial is error, serial is empty ")
	}

	return nil
}

func GetRrdpSnapshot(snapshotUrl string) (snapshotModel SnapshotModel, err error) {

	// 往rp发送请求
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/41896/snapshot.xml"
	belogs.Debug("GetRrdpSnapshot(): snapshotUrl:", snapshotUrl)
	resp, body, err := httpclient.GetHttps(snapshotUrl)
	if err != nil {
		belogs.Error("GetRrdpSnapshot(): snapshotUrl fail, ", snapshotUrl, err)
		return snapshotModel, err
	}
	belogs.Debug("GetRrdpSnapshot(): resp.Status:", resp.Status, "    len(body):", len(body))

	err = xmlutil.UnmarshalXml(body, &snapshotModel)
	if err != nil {
		belogs.Error("GetRrdpSnapshot(): UnmarshalXml fail, ", snapshotUrl, err)
		return snapshotModel, err
	}

	snapshotModel.Hash = hashutil.Sha256([]byte(body))

	return snapshotModel, nil
}

func CheckRrdpSnapshot(snapshotModel *SnapshotModel, notificationModel *NotificationModel) (err error) {
	if snapshotModel.Version != "1" {
		belogs.Error("CheckRrdpSnapshot():  snapshotModel.Version != 1")
		return errors.New("snapshot version is error, version is not 1, it is " + snapshotModel.Version)
	}
	if len(snapshotModel.Session_id) == 0 {
		belogs.Error("CheckRrdpSnapshot(): len(snapshotModel.Session_id) == 0")
		return errors.New("snapshot session_id is error, session_id is empty  ")
	}
	if notificationModel.Session_id != snapshotModel.Session_id {
		belogs.Error("CheckRrdpSnapshot(): snapshotModel.Session_id:", snapshotModel.Session_id,
			"    notificationModel.Session_id:", notificationModel.Session_id)
		return errors.New("snapshot's session_id is different from  notification's session_id")
	}

	if strings.ToLower(notificationModel.Snapshot.Hash) != strings.ToLower(snapshotModel.Hash) {
		belogs.Error("CheckRrdpSnapshot(): snapshotModel.Hash:", snapshotModel.Hash,
			"    notificationModel.Snapshot.Hash:", notificationModel.Snapshot.Hash)
		return errors.New("snapshot's hash is different from  notification's snapshot's hash")
	}
	return nil

}
func GetRrdpDelta(deltaUrl string) (deltaModel DeltaModel, err error) {

	// 往rp发送请求
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/43230/delta.xml"
	belogs.Debug("deltaUrl(): deltaUrl:", deltaUrl)
	resp, body, err := httpclient.GetHttps(deltaUrl)
	if err != nil {
		belogs.Error("deltaUrl(): deltaUrl fail, ", deltaUrl, err)
		return deltaModel, err
	}
	belogs.Debug("deltaUrl(): resp.Status:", resp.Status, "    len(body):", len(body))

	err = xmlutil.UnmarshalXml(body, &deltaModel)
	if err != nil {
		belogs.Error("deltaUrl(): UnmarshalXml fail, ", deltaUrl, err)
		return deltaModel, err
	}

	deltaModel.Hash = hashutil.Sha256([]byte(body))
	return deltaModel, nil
}

func CheckRrdpDelta(deltaModel *DeltaModel, notificationModel *NotificationModel) (err error) {
	if deltaModel.Version != "1" {
		belogs.Error("CheckRrdpDelta():  deltaModel.Version != 1")
		return errors.New("delta version is error, version is not 1, it is " + deltaModel.Version)
	}
	if len(deltaModel.Session_id) == 0 {
		belogs.Error("CheckRrdpDelta(): len(deltaModel.Session_id) == 0")
		return errors.New("delta session_id is error, session_id is empty  ")
	}
	if notificationModel.Session_id != deltaModel.Session_id {
		belogs.Error("CheckRrdpDelta(): deltaModel.Session_id:", deltaModel.Session_id,
			"    notificationModel.Session_id:", notificationModel.Session_id)
		return errors.New("delta's session_id is different from  notification's session_id")
	}
	for i, _ := range deltaModel.DeltaPublishs {
		base64Hash := hashutil.Sha256([]byte(deltaModel.DeltaPublishs[i].Base64))
		if strings.ToLower(base64Hash) != strings.ToLower(deltaModel.DeltaPublishs[i].Hash) {
			belogs.Error("CheckRrdpDelta(): deltaModel.Serial:", deltaModel.Serial,
				"   deltaModel.DeltaPublishs[i].Hash:", deltaModel.DeltaPublishs[i].Hash,
				"    base64Hash:", base64Hash)
			return errors.New("delta's base64's hash is different from  deltaModel's hash")
		}
	}

	if _, ok := notificationModel.MapSerialDeltas[deltaModel.Serial]; !ok {
		belogs.Error("CheckRrdpDelta(): notification has not such  delta's serial:", deltaModel.Serial)
		return errors.New("notification has not such  delta's serial")
	}
	if strings.ToLower(notificationModel.MapSerialDeltas[deltaModel.Serial].Hash) != strings.ToLower(deltaModel.Hash) {
		belogs.Error("CheckRrdpDelta(): deltaModel.Hash:", deltaModel.Hash,
			"    notificationModel.Snapshot.Hash:", notificationModel.Snapshot.Hash)
		return errors.New("delta's hash is different from  notification's snapshot's hash")
	}

	return nil

}
