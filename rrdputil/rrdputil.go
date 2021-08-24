package rrdputil

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cpusoft/goutil/base64util"
	belogs "github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/stringutil"
	"github.com/cpusoft/goutil/urlutil"
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

	// will sort deltas from smaller to bigger
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
	start := time.Now()
	belogs.Debug("getRrdpNotificationImpl(): notificationUrl:", notificationUrl)
	resp, body, err := httpclient.GetHttpsVerify(notificationUrl, true)
	if err == nil {
		defer resp.Body.Close()
		belogs.Debug("getRrdpNotificationImpl(): GetHttpsVerify notificationUrl:", notificationUrl,
			"   resp.Status:", resp.Status, "    len(body):", len(body),
			"   time(s):", time.Now().Sub(start).Seconds())

		if resp.StatusCode != http.StatusOK {
			belogs.Error("getRrdpNotificationImpl(): GetHttpsVerify notificationUrl, is not StatusOK:", notificationUrl,
				"   resp.Status:", resp.Status, "    body:", body)
			return notificationModel, errors.New("http status code of " + notificationUrl + " is " + resp.Status)
		}

	} else {
		belogs.Debug("getRrdpNotificationImpl(): GetHttpsVerify notificationUrl fail, will use curl again:", notificationUrl, "   resp:",
			resp, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)

		// then try using curl
		body, err = httpclient.GetByCurl(notificationUrl)
		if err != nil {
			belogs.Error("getRrdpNotificationImpl(): GetByCurl notificationUrl fail:", notificationUrl, "   resp:",
				resp, "    len(body):", len(body), "       body:", body, "  time(s):", time.Now().Sub(start).Seconds(), err)
			return notificationModel, errors.New("http error of " + notificationUrl + " is " + err.Error())
		}
		belogs.Debug("getRrdpNotificationImpl(): GetByCurl deltaUrl ok", notificationUrl, "    len(body):", len(body),
			"  time(s):", time.Now().Sub(start).Seconds())
	}

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
		belogs.Error("RrdpNotificationTestConnect(): GetHttpsVerify notificationUrl:", notificationUrl, err)
		return errors.New("http error of " + notificationUrl + " is " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		belogs.Error("RrdpNotificationTestConnect(): GetHttpsVerify notificationUrl, is not StatusOK:", notificationUrl,
			"   resp.Status:", resp.Status, "    body:", body)
		return errors.New("http status code of " + notificationUrl + " is " + resp.Status)
	}

	// test is legal
	var notificationModel NotificationModel
	err = xmlutil.UnmarshalXml(body, &notificationModel)
	if err != nil {
		belogs.Error("RrdpNotificationTestConnect(): UnmarshalXml fail: ", notificationUrl, "        body:", body, err)
		return errors.New("response of " + notificationUrl + " is not a legal rrdp file")
	}
	belogs.Info("RrdpNotificationTestConnect(): GetHttpsVerify notificationUrl:", notificationUrl,
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
	resp, body, err := httpclient.GetHttpsVerify(snapshotUrl, true)
	belogs.Debug("getRrdpSnapshotImpl(): GetHttpsVerify, snapshotUrl:", snapshotUrl, "    len(body):", len(body),
		"  time(s):", time.Now().Sub(start).Seconds(), "   err:", err)
	if err == nil {
		defer resp.Body.Close()
		belogs.Debug("getRrdpSnapshotImpl():GetHttpsVerify snapshotUrl ok:", snapshotUrl,
			"    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds())
	} else {
		belogs.Debug("getRrdpSnapshotImpl(): GetHttpsVerify snapshotUrl fail, will use curl again:", snapshotUrl, "   resp:",
			resp, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)

		// then try using curl
		body, err = httpclient.GetByCurl(snapshotUrl)
		if err != nil {
			belogs.Error("getRrdpSnapshotImpl(): GetByCurl snapshotUrl fail:", snapshotUrl, "   resp:",
				resp, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)
			return snapshotModel, err
		}
		belogs.Debug("getRrdpSnapshotImpl(): GetByCurl snapshotUrl ok", snapshotUrl, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds())
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
		belogs.Error("CheckRrdpSnapshot(): snapshotModel.Hash:", snapshotModel.Hash,
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
		pathFileName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, snapshotModel.SnapshotPublishs[i].Uri)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): JoinPrefixPathAndUrlFileName fail:", snapshotModel.SnapshotPublishs[i].Uri)
			return nil, err
		}

		// if dir is notexist ,then mkdir
		dir, file := osutil.Split(pathFileName)
		isExist, err := osutil.IsExists(dir)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): IsExists dir, fail:", dir, err)
			return nil, err
		}

		if !isExist {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				belogs.Error("SaveRrdpSnapshotToRrdpFiles(): MkdirAll dir, fail:", dir, err)
				return nil, err
			}
		}

		bytes, err := base64util.DecodeBase64(strings.TrimSpace(snapshotModel.SnapshotPublishs[i].Base64))
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): DecodeBase64 fail:",
				snapshotModel.SnapshotPublishs[i].Base64, err)
			return nil, err
		}

		err = fileutil.WriteBytesToFile(pathFileName, bytes)
		if err != nil {
			belogs.Error("SaveRrdpSnapshotToRrdpFiles(): WriteBytesToFile fail:", pathFileName,
				len(bytes), err)
			return nil, err
		}
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

func GetRrdpDeltas(notificationModel *NotificationModel, lastSerial uint64) (deltaModels []DeltaModel, err error) {
	start := time.Now()
	belogs.Info("GetRrdpDeltas(): len(notificationModel.Deltas),lastSerial :",
		len(notificationModel.Deltas), lastSerial)

	var wg sync.WaitGroup
	errorMsgCh := make(chan string, len(notificationModel.Deltas))
	deltaModelCh := make(chan DeltaModel, len(notificationModel.Deltas))
	countCh := make(chan int, runtime.NumCPU()*2)
	// serial need from small to large
	for i := len(notificationModel.Deltas) - 1; i >= 0; i-- {
		belogs.Debug("GetRrdpDeltas(): i:", i, "   notificationModel.Deltas[i].Serial:", notificationModel.Deltas[i].Serial)
		if notificationModel.Deltas[i].Serial <= lastSerial {
			belogs.Debug("GetRrdpDeltas():continue, notificationModel.Deltas[i].Serial <= lastSerial:", notificationModel.Deltas[i].Serial, lastSerial)
			continue
		}

		countCh <- 1
		wg.Add(1)
		go getRrdpDeltasImpl(notificationModel, i, deltaModelCh, errorMsgCh, countCh, &wg)
		//
	}
	wg.Wait()
	close(countCh)
	close(errorMsgCh)
	close(deltaModelCh)

	belogs.Debug("GetRrdpDeltas():will get errorMsgCh, and deltaModelCh", len(errorMsgCh), len(deltaModelCh))
	// if has error, then return error
	for errorMsg := range errorMsgCh {
		belogs.Error("GetRrdpDeltas(): getRrdpDeltasImpl fail:", errorMsg, "   time(s):", time.Now().Sub(start).Seconds())
		return nil, errors.New(errorMsg)
	}
	// get deltaModels, and sort
	deltaModels = make([]DeltaModel, 0, len(notificationModel.MapSerialDeltas))
	for deltaModel := range deltaModelCh {
		deltaModels = append(deltaModels, deltaModel)
	}
	sort.Sort(DeltaModelsSort(deltaModels))

	belogs.Info("GetRrdpDeltas():len(deltaModels):", len(deltaModels),
		"   len(notificationModel.Deltas) :", len(notificationModel.Deltas),
		"   lastSerial:", lastSerial, "   time(s):", time.Now().Sub(start).Seconds())

	return deltaModels, nil
}

func getRrdpDeltasImpl(notificationModel *NotificationModel, i int, deltaModelCh chan DeltaModel,
	errorMsgCh chan string, countCh chan int, wg *sync.WaitGroup) {
	defer func() {
		<-countCh
		wg.Done()
	}()

	start := time.Now()
	belogs.Debug("getRrdpDeltasImpl():will notificationModel.Deltas[i].Uri:", i, notificationModel.Deltas[i].Uri)
	deltaModel, err := GetRrdpDelta(notificationModel.Deltas[i].Uri)
	if err != nil {
		belogs.Error("getRrdpDeltasImpl(): GetRrdpDelta fail, delta.Uri :", i,
			notificationModel.Deltas[i].Uri, err, "   time(s):", time.Now().Sub(start).Seconds())
		errorMsgCh <- "get delta " + notificationModel.Deltas[i].Uri + " fail, error is " + err.Error()
		return
	}
	belogs.Debug("getRrdpDeltasImpl():ok notificationModel.Deltas[i].Uri:", i, notificationModel.Deltas[i].Uri,
		"   time(s):", time.Now().Sub(start).Seconds())

	err = CheckRrdpDelta(&deltaModel, notificationModel)
	if err != nil {
		belogs.Error("getRrdpDeltasImpl(): CheckRrdpDelta fail, delta.Uri :", i,
			notificationModel.Deltas[i].Uri, err, "   time(s):", time.Now().Sub(start).Seconds())
		errorMsgCh <- "check delta " + notificationModel.Deltas[i].Uri + " fail, error is " + err.Error()
		return
	}
	belogs.Info("getRrdpDeltasImpl(): delta.Uri:", notificationModel.Deltas[i].Uri,
		"   len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws),
		"   time(s):", time.Now().Sub(start).Seconds())
	deltaModelCh <- deltaModel
	return
}

func GetRrdpDelta(deltaUrl string) (deltaModel DeltaModel, err error) {

	start := time.Now()
	// get delta.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/43230/delta.xml"
	belogs.Debug("GetRrdpDelta(): deltaUrl:", deltaUrl)
	for i := 0; i < 3; i++ {
		deltaModel, err = getRrdpDeltaImpl(deltaUrl)
		if err != nil {
			belogs.Error("GetRrdpDelta():getRrdpDeltaImpl fail, will try again, deltaUrl:", deltaUrl, "  i:", i, err)
		} else {
			break
		}
	}
	if err != nil {
		belogs.Error("GetRrdpDelta():getRrdpDeltaImpl fail:", deltaModel, err)
		return deltaModel, nil
	}

	belogs.Info("GetRrdpDelta(): deltaUrl ok:", deltaUrl, "  time(s):", time.Now().Sub(start).Seconds())
	return deltaModel, nil
}

func getRrdpDeltaImpl(deltaUrl string) (deltaModel DeltaModel, err error) {

	start := time.Now()
	// get delta.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/43230/delta.xml"
	belogs.Debug("getRrdpDeltaImpl(): deltaUrl:", deltaUrl)
	resp, body, err := httpclient.GetHttpsVerify(deltaUrl, true)
	if err == nil {
		defer resp.Body.Close()
		belogs.Debug("getRrdpDeltaImpl(): GetHttpsVerify deltaUrl ok:", deltaUrl, "   resp.Status:",
			resp.Status, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds())
	} else {
		belogs.Debug("getRrdpDeltaImpl(): GetHttpsVerify deltaUrl fail, will use curl again:", deltaUrl, "   resp:",
			resp, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)

		// then try using curl
		body, err = httpclient.GetByCurl(deltaUrl)
		if err != nil {
			belogs.Error("getRrdpDeltaImpl(): GetByCurl deltaUrl fail:", deltaUrl, "   resp:",
				resp, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds(), err)
			return deltaModel, err
		}
		belogs.Debug("getRrdpDeltaImpl(): GetByCurl deltaUrl ok", deltaUrl, "    len(body):", len(body), "  time(s):", time.Now().Sub(start).Seconds())
	}

	err = xmlutil.UnmarshalXml(body, &deltaModel)
	if err != nil {
		belogs.Error("getRrdpDeltaImpl(): UnmarshalXml fail:", deltaUrl, "    body:", body, err)
		return deltaModel, err
	}

	deltaModel.Hash = hashutil.Sha256([]byte(body))
	for i := range deltaModel.DeltaPublishs {
		deltaModel.DeltaPublishs[i].Base64 = stringutil.TrimSpaceAneNewLine(deltaModel.DeltaPublishs[i].Base64)
	}
	deltaModel.DeltaUrl = deltaUrl
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
	/* hash256 comes from the last file
	for i := range deltaModel.DeltaPublishs {
		base64Hash := hashutil.Sha256([]byte((deltaModel.DeltaPublishs[i].Base64)))
		if strings.ToLower(base64Hash) != strings.ToLower(deltaModel.DeltaPublishs[i].Hash) {
			belogs.Error("CheckRrdpDelta(): deltaModel.Serial:", deltaModel.Serial,
				"   deltaModel.DeltaPublishs[i].Hash:", deltaModel.DeltaPublishs[i].Hash,
				"    base64Hash:", base64Hash, "   base64:"+deltaModel.DeltaPublishs[i].Base64+",")
		}
	}
	*/
	if len(notificationModel.MapSerialDeltas) > 0 {
		if _, ok := notificationModel.MapSerialDeltas[deltaModel.Serial]; !ok {
			belogs.Error("CheckRrdpDelta(): notification has not such  delta's serial:", deltaModel.Serial)
			return errors.New("notification has not such  delta's serial")
		}
	}
	/*
		found := false
		for i := range notificationModel.Deltas {
			if notificationModel.Deltas[i].Serial == deltaModel.Serial &&
				strings.ToLower(notificationModel.Deltas[i].Hash) != strings.ToLower(deltaModel.Hash) {
				found = true
				break
			}
		}
		if !found {
			belogs.Error("CheckRrdpDelta(): compare notificationModel.MapSerialDeltas:",
				notificationModel.MapSerialDeltas, "  eltaModel.Serial: ", deltaModel.Serial,
				"    notificationModel.Deltas[i].Hash not found, ",
				"    deltaModel.Hash:", deltaModel.Hash)
			//shaodebug ,not return ,just log error
			//return errors.New("delta's hash is different from  notification's snapshot's hash")
		}
	*/
	return nil

}

// repoPath --> conf.String("rrdp::reporrdp"): /root/rpki/data/reporrdp
func SaveRrdpDeltaToRrdpFiles(deltaModel *DeltaModel, repoPath string) (rrdpFiles []RrdpFile, err error) {

	// delta may have no publishes and no withdraws
	if deltaModel == nil || (len(deltaModel.DeltaPublishs) == 0 && len(deltaModel.DeltaWithdraws) == 0) {
		belogs.Debug("SaveRrdpDeltaToRrdpFiles(): len(snapshotModel.DeltaPublishs)==0 && len(deltaModel.DeltaWithdraws)==0, deltaModel:",
			jsonutil.MarshalJson(deltaModel), "   repoPath:", repoPath)
		return rrdpFiles, nil

	}
	// first , del withdraw files
	for i := range deltaModel.DeltaWithdraws {
		pathFileName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, deltaModel.DeltaWithdraws[i].Uri)
		if err != nil {
			belogs.Error("SaveRrdpDeltaToRrdpFiles(): JoinPrefixPathAndUrlFileName fail:", deltaModel.DeltaWithdraws[i].Uri)
			return nil, err
		}

		// if in this dir, no more files, then del dir
		// will ignore error
		dir, file := osutil.Split(pathFileName)
		files, _ := ioutil.ReadDir(dir)
		belogs.Debug("SaveRrdpDeltaToRrdpFiles():will remove pathFileName:", pathFileName,
			"   dir:", dir, "   files:", len(files))
		os.Remove(pathFileName)
		if len(files) == 0 {
			os.RemoveAll(dir)
		}

		rrdpFile := RrdpFile{
			FilePath:  dir,
			FileName:  file,
			SyncType:  "del",
			SourceUrl: deltaModel.DeltaUrl,
		}
		belogs.Debug("SaveRrdpDeltaToRrdpFiles(): rrdpFile, pathFileName:",
			pathFileName, "   dir:", dir, "   rrdpFile:", jsonutil.MarshalJson(rrdpFile), "  ok")
		rrdpFiles = append(rrdpFiles, rrdpFile)
	}

	// seconde, save publish files
	for i := range deltaModel.DeltaPublishs {
		// get absolute dir /dest/***/***/**.**
		pathFileName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, deltaModel.DeltaPublishs[i].Uri)
		if err != nil {
			belogs.Error("SaveRrdpDeltaToRrdpFiles(): JoinPrefixPathAndUrlFileName fail:", deltaModel.DeltaPublishs[i].Uri)
			return nil, err
		}

		// if dir is notexist ,then mkdir
		dir, file := osutil.Split(pathFileName)
		isExist, err := osutil.IsExists(dir)
		if err != nil {
			belogs.Error("SaveRrdpDeltaToRrdpFiles(): Publish ReadDir fail:", dir, err)
			return nil, err
		}
		if !isExist {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				belogs.Error("SaveRrdpDeltaToRrdpFiles(): Publish MkdirAll fail:", dir, err)
				return nil, err
			}
		}

		// decode base65 to bytes
		bytes, err := base64util.DecodeBase64(strings.TrimSpace(deltaModel.DeltaPublishs[i].Base64))
		if err != nil {
			belogs.Error("SaveRrdpDeltaToRrdpFiles():Publish DecodeBase64 fail:",
				deltaModel.Serial,
				deltaModel.DeltaPublishs[i].Uri, deltaModel.DeltaPublishs[i].Base64, err)
			return nil, err
		}

		err = fileutil.WriteBytesToFile(pathFileName, bytes)
		if err != nil {
			belogs.Error("SaveRrdpDeltaToRrdpFiles():Publish WriteBytesToFile fail:",
				deltaModel.Serial,
				deltaModel.DeltaPublishs[i].Uri,
				pathFileName, len(bytes), err)
			return nil, err
		}

		// some rrdp have no withdraw, only publish, so change to update to del old in db
		rrdpFile := RrdpFile{
			FilePath: dir,
			FileName: file,
			//SyncType: "add",
			SyncType:  "update",
			SourceUrl: deltaModel.DeltaUrl,
		}
		belogs.Debug("SaveRrdpDeltaToRrdpFiles():Publish Remove rrdpFile ", jsonutil.MarshalJson(rrdpFile), "  ok")
		rrdpFiles = append(rrdpFiles, rrdpFile)
	}
	belogs.Info("SaveRrdpSnapshotToRrdpFiles(): save len(rrdpFiles): ", len(rrdpFiles))
	belogs.Debug("SaveRrdpSnapshotToRrdpFiles(): save rrdpFiles: ", jsonutil.MarshalJson(rrdpFiles))
	return rrdpFiles, nil

}
