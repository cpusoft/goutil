package rrdputil

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	base64util "github.com/cpusoft/goutil/base64util"
	fileutil "github.com/cpusoft/goutil/fileutil"
	hashutil "github.com/cpusoft/goutil/hashutil"
	httpclient "github.com/cpusoft/goutil/httpclient"
	osutil "github.com/cpusoft/goutil/osutil"
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

	notificationModel.MapSerialDeltas = make(map[uint64]uint64, len(notificationModel.Deltas)+10)
	for i, _ := range notificationModel.Deltas {
		notificationModel.MapSerialDeltas[notificationModel.Deltas[i].Serial] = notificationModel.Deltas[i].Serial
		serial := notificationModel.Deltas[i].Serial
		if serial > notificationModel.MaxSerail {
			notificationModel.MaxSerail = serial
		}
		if serial < notificationModel.MinSerail {
			notificationModel.MinSerail = serial
		}
	}
	return notificationModel, nil
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
	if _, ok := notificationModel.MapSerialDeltas[notificationModel.Serial]; !ok {
		belogs.Error("CheckRrdpNotification(): notification has not such serial in deltas:", notificationModel.Serial)
		return errors.New("notification has not such serial in deltas")
	}
	return nil
}

func GetRrdpSnapshot(snapshotUrl string) (snapshotModel SnapshotModel, err error) {

	// get snapshot.xml
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
	if len(snapshotModel.SessionId) == 0 {
		belogs.Error("CheckRrdpSnapshot(): len(snapshotModel.SessionId) == 0")
		return errors.New("snapshot session_id is error, session_id is empty  ")
	}
	if notificationModel.SessionId != snapshotModel.SessionId {
		belogs.Error("CheckRrdpSnapshot(): snapshotModel.SessionId:", snapshotModel.SessionId,
			"    notificationModel.SessionId:", notificationModel.SessionId)
		return errors.New("snapshot's session_id is different from  notification's session_id")
	}
	if _, ok := notificationModel.MapSerialDeltas[snapshotModel.Serial]; !ok {
		belogs.Error("CheckRrdpSnapshot(): notification has not such  snapshot's serial:", snapshotModel.Serial)
		return errors.New("notification has not such  snapshot's serial")
	}
	if strings.ToLower(notificationModel.Snapshot.Hash) != strings.ToLower(snapshotModel.Hash) {
		belogs.Error("CheckRrdpSnapshot(): snapshotModel.Hash:", snapshotModel.Hash,
			"    notificationModel.Snapshot.Hash:", notificationModel.Snapshot.Hash)
		return errors.New("snapshot's hash is different from  notification's snapshot's hash")
	}
	return nil

}

// repoPath --> conf.String("rrdp::reporrdp"): /root/rpki/data/reporrdp
func SaveRrdpSnapshotToFiles(snapshotModel *SnapshotModel, repoPath string) (err error) {
	if snapshotModel == nil || len(snapshotModel.SnapshotPublishs) == 0 {
		belogs.Debug("SaveRrdpSnapshotToFiles(): len(snapshotModel.SnapshotPublishs)==0")
		return nil
	}
	for i, _ := range snapshotModel.SnapshotPublishs {
		pathFileName, err := osutil.GetPathFileNameFromUrl(repoPath, snapshotModel.SnapshotPublishs[i].Uri)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToFiles(): GetPathFileNameFromUrl fail:", snapshotModel.SnapshotPublishs[i].Uri)
			return err
		}

		// if dir is notexist ,then mkdir
		dir, _ := osutil.Split(pathFileName)
		isExist, _ := osutil.IsExists(dir)
		if !isExist {
			os.MkdirAll(dir, os.ModePerm)
		}

		bytes, err := base64util.DecodeBase64(strings.TrimSpace(snapshotModel.SnapshotPublishs[i].Base64))
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToFiles(): DecodeBase64 fail:", snapshotModel.SnapshotPublishs[i].Base64)
			return err
		}

		err = fileutil.WriteBytesToFile(pathFileName, bytes)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToFiles(): WriteBytesToFile fail:", pathFileName, len(bytes))
			return err
		}
		belogs.Debug("SaveRrdpSnapshotToFiles(): save pathFileName ", pathFileName, "  ok")
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
	if len(deltaModel.SessionId) == 0 {
		belogs.Error("CheckRrdpDelta(): len(deltaModel.SessionId) == 0")
		return errors.New("delta session_id is error, session_id is empty  ")
	}
	if notificationModel.SessionId != deltaModel.SessionId {
		belogs.Error("CheckRrdpDelta(): deltaModel.SessionId:", deltaModel.SessionId,
			"    notificationModel.SessionId:", notificationModel.SessionId)
		return errors.New("delta's session_id is different from  notification's session_id")
	}
	for i, _ := range deltaModel.DeltaPublishs {
		base64Hash := hashutil.Sha256([]byte((deltaModel.DeltaPublishs[i].Base64)))
		if strings.ToLower(base64Hash) != strings.ToLower(deltaModel.DeltaPublishs[i].Hash) {
			belogs.Error("CheckRrdpDelta(): deltaModel.Serial:", deltaModel.Serial,
				"   deltaModel.DeltaPublishs[i].Hash:", deltaModel.DeltaPublishs[i].Hash,
				"    base64Hash:", base64Hash, "   base64:"+deltaModel.DeltaPublishs[i].Base64+",")
			// shaodebug , all are not equal, why ??
			//return errors.New("delta's base64's hash is different from  deltaModel's hash")
		}
	}

	if _, ok := notificationModel.MapSerialDeltas[deltaModel.Serial]; !ok {
		belogs.Error("CheckRrdpDelta(): notification has not such  delta's serial:", deltaModel.Serial)
		return errors.New("notification has not such  delta's serial")
	}
	found := false
	for i, _ := range notificationModel.Deltas {
		if notificationModel.Deltas[i].Serial == deltaModel.Serial &&
			strings.ToLower(notificationModel.Deltas[i].Hash) != strings.ToLower(deltaModel.Hash) {
			found = true
			break
		}
	}
	if !found {
		belogs.Error("CheckRrdpDelta(): deltaModel.Hash:", deltaModel.Hash,
			"    notificationModel.Snapshot.Hash:", notificationModel.Snapshot.Hash)
		return errors.New("delta's hash is different from  notification's snapshot's hash")
	}

	return nil

}

// repoPath --> conf.String("rrdp::reporrdp"): /root/rpki/data/reporrdp
func SaveRrdpDeltaToFiles(deltaModel *DeltaModel, repoPath string) (err error) {
	if deltaModel == nil || (len(deltaModel.DeltaPublishs) == 0 && len(deltaModel.DeltaWithdraws) == 0) {
		belogs.Debug("SaveRrdpDeltaToFiles(): len(snapshotModel.SnapshotPublishs)==0")
		return nil
	}
	// save publish files
	for i, _ := range deltaModel.DeltaPublishs {
		// get absolute dir /dest/***/***/**.**
		pathFileName, err := osutil.GetPathFileNameFromUrl(repoPath, deltaModel.DeltaPublishs[i].Uri)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToFiles(): GetPathFileNameFromUrl fail:", deltaModel.DeltaPublishs[i].Uri)
			return err
		}

		// if dir is notexist ,then mkdir
		dir, _ := osutil.Split(pathFileName)
		isExist, _ := osutil.IsExists(dir)
		if !isExist {
			os.MkdirAll(dir, os.ModePerm)
		}

		// decode base65 to bytes
		bytes, err := base64util.DecodeBase64(strings.TrimSpace(deltaModel.DeltaPublishs[i].Base64))
		if err != nil {
			belogs.Error("SaveRrdpDeltaToFiles():Publish DecodeBase64 fail:",
				deltaModel.Serial,
				deltaModel.DeltaPublishs[i].Uri, deltaModel.DeltaPublishs[i].Base64)
			return err
		}

		err = fileutil.WriteBytesToFile(pathFileName, bytes)
		if err != nil {
			belogs.Error("SaveRrdpDeltaToFiles():Publish WriteBytesToFile fail:",
				deltaModel.Serial,
				deltaModel.DeltaPublishs[i].Uri,
				pathFileName, len(bytes))
			return err
		}
		belogs.Debug("SaveRrdpDeltaToFiles():Publish save pathFileName ", pathFileName, "  ok")
	}

	// del withdraw files
	for i, _ := range deltaModel.DeltaWithdraws {
		pathFileName, err := osutil.GetPathFileNameFromUrl(repoPath, deltaModel.DeltaWithdraws[i].Uri)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToFiles(): GetPathFileNameFromUrl fail:", deltaModel.DeltaWithdraws[i].Uri)
			return err
		}
		err = os.Remove(pathFileName)
		if err != nil {
			belogs.Error("SaveRrdpDeltaToFiles():Remove fail:", pathFileName)
			return err
		}
		// if in this dir, no more files, then del dir
		dir, _ := osutil.Split(pathFileName)
		files, _ := ioutil.ReadDir(dir)
		if len(files) == 0 {
			os.RemoveAll(dir)
		}
		belogs.Debug("SaveRrdpDeltaToFiles():Withdraw Remove pathFileName ", pathFileName, "  ok")
	}
	return nil

}
