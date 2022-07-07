package rrdputil

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/cpusoft/goutil/base64util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/netutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/urlutil"
	"github.com/cpusoft/goutil/xmlutil"
)

func GetRrdpSnapshot(snapshotUrl string) (snapshotModel SnapshotModel, err error) {
	start := time.Now()
	// get snapshot.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/41896/snapshot.xml"
	belogs.Info("GetRrdpSnapshot():will get snapshotUrl:", snapshotUrl)
	for i := 0; i < 3; i++ {
		snapshotModel, err = getRrdpSnapshotImpl(snapshotUrl)
		if err != nil {
			belogs.Error("GetRrdpSnapshot():getRrdpSnapshotImpl fail, will try again, snapshotUrl:", snapshotUrl, "  i:", i, err)
		} else {
			break
		}
	}
	if err != nil {
		belogs.Error("GetRrdpSnapshot():getRrdpSnapshotImpl fail:", snapshotUrl, err)
		return snapshotModel, nil
	}
	belogs.Info("GetRrdpSnapshot(): snapshotUrl ok:", snapshotUrl, "  time(s):", time.Now().Sub(start).Seconds())
	return snapshotModel, nil
}

func getRrdpSnapshotImpl(snapshotUrl string) (snapshotModel SnapshotModel, err error) {
	start := time.Now()
	// get snapshot.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/41896/snapshot.xml"
	belogs.Debug("getRrdpSnapshotImpl(): snapshotUrl:", snapshotUrl)
	snapshotUrl = strings.TrimSpace(snapshotUrl)
	resp, body, err := httpclient.GetHttpsVerify(snapshotUrl, true)
	belogs.Debug("getRrdpSnapshotImpl(): GetHttpsVerify, snapshotUrl:", snapshotUrl, "   ipAddrs:", netutil.LookupIpByUrl(snapshotUrl),
		"    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), "   err:", err)
	if err == nil {
		defer resp.Body.Close()
		belogs.Debug("getRrdpSnapshotImpl():GetHttpsVerify snapshotUrl ok:", snapshotUrl,
			"   ipAddrs:", netutil.LookupIpByUrl(snapshotUrl),
			"   len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds())
	} else {
		belogs.Debug("getRrdpSnapshotImpl(): GetHttpsVerify snapshotUrl fail, will use curl again:", snapshotUrl, "   resp:",
			resp, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)

		// then try using curl
		body, err = httpclient.GetByCurl(snapshotUrl)
		if err != nil {
			belogs.Error("getRrdpSnapshotImpl(): GetByCurl snapshotUrl fail:", snapshotUrl,
				"   ipAddrs:", netutil.LookupIpByUrl(snapshotUrl), "   resp:", resp,
				"   len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)
			return snapshotModel, err
		}
		belogs.Debug("getRrdpSnapshotImpl(): GetByCurl snapshotUrl ok", snapshotUrl, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds())
	}
	// check if body is xml file
	if !strings.Contains(body, `<snapshot`) {
		belogs.Error("GetRrdpSnapshot(): body is not xml file:", snapshotUrl, "   resp:",
			resp, "    len(body):", len(body), "       body:", body, "  time(s):", time.Now().Sub(start).Seconds(), err)
		return snapshotModel, errors.New("body of " + snapshotUrl + " is not xml")
	}

	// get snapshotModel
	err = xmlutil.UnmarshalXml(body, &snapshotModel)
	if err != nil {
		belogs.Error("GetRrdpSnapshot(): UnmarshalXml fail:", snapshotUrl, "   body:", body, err)
		return snapshotModel, err
	}
	snapshotModel.Hash = hashutil.Sha256([]byte(body))
	snapshotModel.SnapshotUrl = snapshotUrl
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
	if len(notificationModel.MapSerialDeltas) > 0 {
		if _, ok := notificationModel.MapSerialDeltas[snapshotModel.Serial]; !ok {
			belogs.Error("CheckRrdpSnapshot(): notification has not such  snapshot's serial:", snapshotModel.Serial)
			return errors.New("notification has not such  snapshot's serial")
		}
	}
	if strings.ToLower(notificationModel.Snapshot.Hash) != strings.ToLower(snapshotModel.Hash) {
		belogs.Info("CheckRrdpSnapshot(): snapshotModel.Hash:", snapshotModel.Hash,
			"    notificationModel.Snapshot.Hash:", notificationModel.Snapshot.Hash, " but just continue")
		//return errors.New("snapshot's hash is different from  notification's snapshot's hash")
	}
	return nil

}

// repoPath --> conf.String("rrdp::reporrdp"): /root/rpki/data/reporrdp
func SaveRrdpSnapshotToRrdpFiles(snapshotModel *SnapshotModel, repoPath string) (rrdpFiles []RrdpFile, err error) {
	if snapshotModel == nil || len(snapshotModel.SnapshotPublishs) == 0 {
		belogs.Error("SaveRrdpSnapshotToRrdpFiles(): len(snapshotModel.SnapshotPublishs)==0")
		return nil, errors.New("snapshot's publishs is empty")
	}
	for i := range snapshotModel.SnapshotPublishs {
		uri := strings.Replace(snapshotModel.SnapshotPublishs[i].Uri, "../", "/", -1) //fix Path traversal
		belogs.Debug("SaveRrdpSnapshotToRrdpFiles():snapshotModel.SnapshotPublishs[i].Uri:", snapshotModel.SnapshotPublishs[i].Uri,
			" uri:", uri)
		pathFileName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, uri)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): JoinPrefixPathAndUrlFileName fail:", snapshotModel.SnapshotPublishs[i].Uri,
				"    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, err)
			return nil, err
		}

		// if dir is notexist ,then mkdir
		dir, file := osutil.Split(pathFileName)
		if !fileutil.CheckPathNameMaxLength(dir) {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): CheckPathNameMaxLength fail,dir:", dir)
			return nil, errors.New("snapshot path name is too long")
		}
		if !fileutil.CheckFileNameMaxLength(file) {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): CheckFileNameMaxLength fail,file:", file)
			return nil, errors.New("snapshot file name is too long")
		}

		isExist, err := osutil.IsExists(dir)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): IsExists dir, fail:", dir, err)
			return nil, err
		}

		if !isExist {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				belogs.Error("SaveRrdpSnapshotToRrdpFiles(): MkdirAll dir, fail:", dir, "    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, err)
				return nil, err
			}
		}

		bytes, err := base64util.DecodeBase64(strings.TrimSpace(snapshotModel.SnapshotPublishs[i].Base64))
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): DecodeBase64 fail:",
				snapshotModel.SnapshotPublishs[i].Base64,
				"    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, err)
			return nil, err
		}

		err = fileutil.WriteBytesToFile(pathFileName, bytes)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): WriteBytesToFile fail:", pathFileName,
				len(bytes), "    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, err)
			return nil, err
		}
		belogs.Info("SaveRrdpSnapshotToRrdpFiles(): update pathFileName:", pathFileName,
			"    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, "  ok")
		belogs.Debug("SaveRrdpSnapshotToRrdpFiles(): save pathFileName ", pathFileName, "  ok")

		rrdpFile := RrdpFile{
			FilePath:  dir,
			FileName:  file,
			SyncType:  "add",
			SourceUrl: snapshotModel.SnapshotUrl,
		}
		rrdpFiles = append(rrdpFiles, rrdpFile)
	}
	belogs.Debug("SaveRrdpSnapshotToRrdpFiles(): save rrdpFiles ", jsonutil.MarshalJson(rrdpFiles))
	return rrdpFiles, nil

}
