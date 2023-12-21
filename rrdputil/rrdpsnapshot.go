package rrdputil

import (
	"errors"
	"net/http"
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
	"github.com/cpusoft/goutil/stringutil"
	"github.com/cpusoft/goutil/urlutil"
	"github.com/cpusoft/goutil/xmlutil"
)

func GetRrdpSnapshot(snapshotUrl string) (snapshotModel SnapshotModel, err error) {
	// get snapshot.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/41896/snapshot.xml"
	belogs.Info("GetRrdpSnapshot():will get snapshotUrl:", snapshotUrl)
	return GetRrdpSnapshotWithConfig(snapshotUrl, nil)
}

func GetRrdpSnapshotWithConfig(snapshotUrl string, httpClientConfig *httpclient.HttpClientConfig) (snapshotModel SnapshotModel, err error) {
	start := time.Now()
	// get snapshot.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/41896/snapshot.xml"
	if httpClientConfig == nil {
		httpClientConfig = httpclient.CloneGLobalHttpClient()
	}
	belogs.Info("GetRrdpSnapshotWithConfig():will get snapshotUrl:", snapshotUrl, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))

	snapshotModel, err = getRrdpSnapshotImplWithConfig(snapshotUrl, httpClientConfig)
	if err != nil {
		belogs.Error("GetRrdpSnapshotWithConfig():getRrdpSnapshotImpl fail:", snapshotUrl, err)
		return snapshotModel, err
	}
	belogs.Info("GetRrdpSnapshotWithConfig(): snapshotUrl ok:", snapshotUrl, "  time(s):", time.Since(start))
	return snapshotModel, nil
}

func getRrdpSnapshotImplWithConfig(snapshotUrl string, httpClientConfig *httpclient.HttpClientConfig) (snapshotModel SnapshotModel, err error) {

	// get snapshot.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/41896/snapshot.xml"
	belogs.Debug("getRrdpSnapshotImplWithConfig(): snapshotUrl:", snapshotUrl, "  httpClientConfig:", jsonutil.MarshalJson(*httpClientConfig))
	snapshotUrl = strings.TrimSpace(snapshotUrl)
	start := time.Now()
	httpclient.SetTimeout(30) // 30mins
	defer httpclient.ResetTimeout()
	resp, body, err := httpclient.GetHttpsVerifyWithConfig(snapshotUrl, true, httpClientConfig)
	belogs.Debug("getRrdpSnapshotImplWithConfig(): GetHttpsVerify, snapshotUrl:", snapshotUrl, "   ipAddrs:", netutil.LookupIpByUrl(snapshotUrl),
		"    len(body):", len(body), "  time(s):", time.Since(start), "   err:", err)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			belogs.Error("getRrdpSnapshotImplWithConfig(): GetHttpsVerify snapshotUrl, is not StatusOK:", snapshotUrl,
				"   resp.Status:", resp.Status, "    body:", stringutil.OmitString(body, 100))
			return snapshotModel, errors.New("http status code of " + snapshotUrl + " is " + resp.Status)
		} else {
			belogs.Debug("getRrdpSnapshotImplWithConfig():GetHttpsVerify snapshotUrl ok:", snapshotUrl,
				"   ipAddrs:", netutil.LookupIpByUrl(snapshotUrl),
				"   len(body):", len(body), "  time(s):", time.Since(start))
		}
	} else {
		belogs.Debug("getRrdpSnapshotImplWithConfig(): GetHttpsVerify snapshotUrl fail, will use curl again:", snapshotUrl, "   resp:",
			resp, "    len(body):", len(body), "  time(s):", time.Since(start), err)

		// then try using curl, using ipv4
		// httpclient.NewHttpClientConfigWithParam(30, 3, "ipv4")
		start = time.Now()
		httpClientConfig.IpType = "ipv4"
		body, err = httpclient.GetByCurlWithConfig(snapshotUrl, httpClientConfig)
		if err != nil {
			belogs.Debug("getRrdpSnapshotImplWithConfig(): GetByCurlWithConfig snapshotUrl, iptype is ipv4, fail:", snapshotUrl,
				"   ipAddrs:", netutil.LookupIpByUrl(snapshotUrl), "   resp:", resp,
				"   len(body):", len(body), "  time(s):", time.Since(start), err)

			// then try again using curl, using all
			start = time.Now()
			httpClientConfig.IpType = "all"
			body, err = httpclient.GetByCurlWithConfig(snapshotUrl, httpClientConfig)
			if err != nil {
				belogs.Error("getRrdpSnapshotImplWithConfig(): GetByCurlWithConfig snapshotUrl, iptype is all, fail:", snapshotUrl,
					"   ipAddrs:", netutil.LookupIpByUrl(snapshotUrl), "   resp:", resp,
					"   len(body):", len(body), "  time(s):", time.Since(start), err)
				return snapshotModel, errors.New("http error of " + snapshotUrl + " is " + err.Error())
			}
			belogs.Debug("getRrdpSnapshotImplWithConfig(): GetByCurlWithConfig snapshotUrl, iptype is all, ok", snapshotUrl, "    len(body):", len(body), "  time(s):", time.Since(start))
		} else {
			belogs.Debug("getRrdpSnapshotImplWithConfig(): GetByCurlWithConfig snapshotUrl, iptype is ipv4, ok", snapshotUrl, "    len(body):", len(body), "  time(s):", time.Since(start))
		}
	}
	// check if body is xml file
	if !strings.Contains(body, `<snapshot`) {
		belogs.Error("getRrdpSnapshotImplWithConfig(): body is not xml file:", snapshotUrl, "   resp:",
			resp, "    len(body):", len(body), "       body:", stringutil.OmitString(body, 100), "  time(s):", time.Since(start), err)
		return snapshotModel, errors.New("body of " + snapshotUrl + " is not xml")
	}

	// get snapshotModel
	belogs.Debug("getRrdpSnapshotImplWithConfig(): get body, snapshotUrl:", snapshotUrl, " len(body):", len(body))
	err = xmlutil.UnmarshalXml(body, &snapshotModel)
	if err != nil {
		belogs.Error("getRrdpSnapshotImplWithConfig(): UnmarshalXml fail:", snapshotUrl, "    len(body):", len(body),
			"   body:", stringutil.OmitString(body, 100), err)
		return snapshotModel, err
	}
	belogs.Debug("getRrdpSnapshotImplWithConfig(): len(snapshotModel.SnapshotPublishs):", len(snapshotModel.SnapshotPublishs))
	for i := range snapshotModel.SnapshotPublishs {
		uri := strings.Replace(snapshotModel.SnapshotPublishs[i].Uri, "../", "/", -1) //fix Path traversal
		snapshotModel.SnapshotPublishs[i].Uri = uri
		base64 := base64util.TrimBase64(snapshotModel.SnapshotPublishs[i].Base64)
		snapshotModel.SnapshotPublishs[i].Base64 = base64
	}
	snapshotModel.Hash = hashutil.Sha256([]byte(body))
	snapshotModel.SnapshotUrl = snapshotUrl

	belogs.Info("getRrdpSnapshotImplWithConfig(): get from snapshotUrl ok", snapshotUrl,
		"   len(snapshotModel.SnapshotPublishs):", len(snapshotModel.SnapshotPublishs), "  time(s):", time.Since(start))
	return snapshotModel, nil
}

func CheckRrdpSnapshot(snapshotModel *SnapshotModel, notificationModel *NotificationModel) (err error) {
	err = CheckRrdpSnapshotValue(snapshotModel)
	if err != nil {
		belogs.Error("CheckRrdpSnapshot():  CheckRrdpSnapshotValue fail:", err)
		return err
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
func CheckRrdpSnapshotValue(snapshotModel *SnapshotModel) error {
	if snapshotModel.Version != "1" {
		belogs.Error("CheckRrdpSnapshotValue():  snapshotModel.Version != 1. current snapshot version is outdated, url is " + snapshotModel.SnapshotUrl)
		return errors.New("current snapshot version is not one, url is " + snapshotModel.SnapshotUrl)
	}
	if len(snapshotModel.SessionId) == 0 {
		belogs.Error("CheckRrdpSnapshotValue(): len(snapshotModel.SessionId) == 0")
		return errors.New("snapshot session id is empty, url is " + snapshotModel.SnapshotUrl)
	}
	return nil
}

// repoPath --> conf.String("rrdp::reporrdp"): /root/rpki/data/reporrdp
func SaveRrdpSnapshotToRrdpFiles(snapshotModel *SnapshotModel, repoPath string) (rrdpFiles []RrdpFile, err error) {
	if snapshotModel == nil || len(snapshotModel.SnapshotPublishs) == 0 {
		belogs.Error("SaveRrdpSnapshotToRrdpFiles(): len(snapshotModel.SnapshotPublishs)==0")
		return nil, errors.New("snapshot's publishs is empty")
	}
	duplicateFilePathNames := make(map[string]string, 0)
	for i := range snapshotModel.SnapshotPublishs {
		uri := strings.Replace(snapshotModel.SnapshotPublishs[i].Uri, "../", "/", -1) //fix Path traversal
		belogs.Debug("SaveRrdpSnapshotToRrdpFiles():snapshotModel.SnapshotPublishs[i].Uri:", snapshotModel.SnapshotPublishs[i].Uri,
			" uri:", uri)
		filePathName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, uri)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): JoinPrefixPathAndUrlFileName fail:", snapshotModel.SnapshotPublishs[i].Uri,
				"    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, err)
			return nil, err
		}

		// if dir is notexist ,then mkdir
		dir, file := osutil.Split(filePathName)
		if !fileutil.CheckPathNameMaxLength(dir) {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): CheckPathNameMaxLength fail,dir:", dir)
			return nil, errors.New("snapshot path name is too long")
		}
		if !fileutil.CheckFileNameMaxLength(file) {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): CheckFileNameMaxLength fail,file:", file)
			return nil, errors.New("snapshot file name is too long")
		}

		if _, ok := duplicateFilePathNames[filePathName]; ok {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): duplicate file in snapshot, fail, filePathName:", filePathName,
				"   snapshotModel:", snapshotModel.String())
			continue
		} else {
			duplicateFilePathNames[filePathName] = filePathName
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

		err = fileutil.WriteBytesToFile(filePathName, bytes)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): WriteBytesToFile fail:", filePathName,
				len(bytes), "    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, err)
			return nil, err
		}
		belogs.Info("SaveRrdpSnapshotToRrdpFiles(): update filePathName:", filePathName,
			"    snapshotModel.SnapshotUrl:", snapshotModel.SnapshotUrl, "  ok")
		belogs.Debug("SaveRrdpSnapshotToRrdpFiles(): save filePathName ", filePathName, "  ok")

		rrdpFile := RrdpFile{
			FilePath:  dir,
			FileName:  file,
			SyncType:  "add",
			SourceUrl: snapshotModel.SnapshotUrl,
		}
		rrdpFiles = append(rrdpFiles, rrdpFile)
	}
	belogs.Debug("SaveRrdpSnapshotToRrdpFiles(): save len(rrdpFiles):", len(rrdpFiles))
	return rrdpFiles, nil

}
