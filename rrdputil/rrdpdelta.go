package rrdputil

import (
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
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

func GetRrdpDeltas(notificationModel *NotificationModel, lastSerial uint64) (deltaModels []DeltaModel, err error) {
	start := time.Now()
	belogs.Info("GetRrdpDeltas(): len(notificationModel.Deltas),lastSerial :",
		len(notificationModel.Deltas), lastSerial)

	var wg sync.WaitGroup
	errorMsgCh := make(chan string, len(notificationModel.Deltas))
	deltaModelCh := make(chan DeltaModel, len(notificationModel.Deltas))
	countCh := make(chan int, runtime.NumCPU()*2)
	// serial need from newest to oldest
	for i := 0; i < len(notificationModel.Deltas); i++ {
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
		belogs.Error("GetRrdpDeltas(): getRrdpDeltasImpl fail:", errorMsg, "   time(s):", time.Since(start))
		return nil, errors.New(errorMsg)
	}
	// get deltaModels, and sort
	deltaModels = make([]DeltaModel, 0, len(notificationModel.MapSerialDeltas))
	for deltaModel := range deltaModelCh {
		deltaModels = append(deltaModels, deltaModel)
	}
	// sort, from newest to oldest
	sort.Sort(DeltaModelsSort(deltaModels))

	belogs.Info("GetRrdpDeltas():len(deltaModels):", len(deltaModels),
		"   len(notificationModel.Deltas) :", len(notificationModel.Deltas),
		"   lastSerial:", lastSerial, "   time(s):", time.Since(start))

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
			notificationModel.Deltas[i].Uri, err, "   time(s):", time.Since(start))
		errorMsgCh <- "get delta " + notificationModel.Deltas[i].Uri + " fail, error is " + err.Error()
		return
	}
	belogs.Debug("getRrdpDeltasImpl():ok notificationModel.Deltas[i].Uri:", i, notificationModel.Deltas[i].Uri,
		"   time(s):", time.Since(start))

	err = CheckRrdpDelta(&deltaModel, notificationModel)
	if err != nil {
		belogs.Error("getRrdpDeltasImpl(): CheckRrdpDelta fail, delta.Uri :", i,
			notificationModel.Deltas[i].Uri, err, "   time(s):", time.Since(start))
		errorMsgCh <- "check delta " + notificationModel.Deltas[i].Uri + " fail, error is " + err.Error()
		return
	}
	belogs.Info("getRrdpDeltasImpl(): delta.Uri:", notificationModel.Deltas[i].Uri,
		"   len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws),
		"   time(s):", time.Since(start))
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
		belogs.Error("GetRrdpDelta():getRrdpDeltaImpl fail:", deltaUrl, err)
		return deltaModel, err
	}

	belogs.Info("GetRrdpDelta(): deltaUrl ok:", deltaUrl, "  time(s):", time.Since(start))
	return deltaModel, nil
}

func getRrdpDeltaImpl(deltaUrl string) (deltaModel DeltaModel, err error) {

	// get delta.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/43230/delta.xml"
	belogs.Debug("getRrdpDeltaImpl(): deltaUrl:", deltaUrl)
	deltaUrl = strings.TrimSpace(deltaUrl)
	start := time.Now()
	resp, body, err := httpclient.GetHttpsVerify(deltaUrl, true)
	if err == nil {
		defer resp.Body.Close()
		belogs.Debug("getRrdpDeltaImpl(): GetHttpsVerify deltaUrl ok:", deltaUrl, "   resp.Status:", resp.Status,
			"   ipAddrs:", netutil.LookupIpByUrl(deltaUrl),
			"   len(body):", len(body), "  time(s):", time.Since(start))
	} else {
		belogs.Error("getRrdpDeltaImpl(): GetHttpsVerify deltaUrl fail, will use curl again:", deltaUrl, "   ipAddrs:", netutil.LookupIpByUrl(deltaUrl),
			"   resp:", resp, "    len(body):", len(body), "  time(s):", time.Since(start), err)

		// then try using curl
		start = time.Now()
		body, err = httpclient.GetByCurl(deltaUrl)
		if err != nil {
			belogs.Error("getRrdpDeltaImpl(): GetByCurl deltaUrl fail:", deltaUrl, "   resp:", resp,
				"   ipAddrs:", netutil.LookupIpByUrl(deltaUrl),
				"   len(body):", len(body), "  time(s):", time.Since(start), err)
			return deltaModel, err
		}
		belogs.Debug("getRrdpDeltaImpl(): GetByCurl deltaUrl ok", deltaUrl, "    len(body):", len(body), "  time(s):", time.Since(start))
	}

	// check if body is xml file
	if !strings.Contains(body, `<delta`) {
		belogs.Error("GetRrdpSnapshot(): body is not xml file:", deltaUrl, "   resp:",
			resp, "    len(body):", len(body), "       body:", body, "  time(s):", time.Since(start), err)
		return deltaModel, errors.New("body of " + deltaUrl + " is not xml")
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

	for i := range notificationModel.Deltas {
		if notificationModel.Deltas[i].Serial == deltaModel.Serial {
			if deltaModel.Hash != notificationModel.Deltas[i].Hash {
				belogs.Info("CheckRrdpDelta(): deltaModel.Hash is not equal to notificationModel.Deltas[i].Hash,",
					"   deltaModel.Serial:", deltaModel.Serial, "    deltaModel.Hash:", deltaModel.Hash,
					"   notificationModel.Deltas[i].Hash:", notificationModel.Deltas[i].Hash, " but just continue")
			}
		}
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

	return nil

}

func SaveRrdpDeltasToRrdpFiles(deltaModels []DeltaModel, notifyUrl, destPath string) (rrdpFilesAll []RrdpFile, err error) {

	rrdpFilesAll = make([]RrdpFile, 0)
	rrdpUris := make(map[string]uint64, len(deltaModels)+20)
	// from latest to oldest
	// will use latest serial delta, and ignore same url files in older serial delta
	for i := range deltaModels {
		// save publish files and remove withdraw files
		rrdpFiles, err := saveRrdpDeltaToRrdpFiles(&deltaModels[i], rrdpUris, destPath)
		if err != nil {
			belogs.Error("processRrdpDelta(): saveRrdpDeltaToRrdpFiles fail, notifyUrl:", notifyUrl,
				"   deltaModels[i].SessionId:", deltaModels[i].SessionId,
				"   deltaModels[i].Serial:", deltaModels[i].Serial, "   deltaModels[i].DeltaUrl:", deltaModels[i].DeltaUrl,
				"   snapshotDeltaResult.DestPath: ", destPath, err)
			return nil, err
		}
		// add to head
		rrdpFilesAll = append(rrdpFiles, rrdpFilesAll...)
	}
	return rrdpFilesAll, nil
}

// repoPath --> conf.String("rrdp::reporrdp"): /root/rpki/data/reporrdp
func saveRrdpDeltaToRrdpFiles(deltaModel *DeltaModel, rrdpUris map[string]uint64, repoPath string) (rrdpFiles []RrdpFile, err error) {

	// delta may have no publishes and no withdraws
	if deltaModel == nil || (len(deltaModel.DeltaPublishs) == 0 && len(deltaModel.DeltaWithdraws) == 0) {
		belogs.Debug("saveRrdpDeltaToRrdpFiles(): len(snapshotModel.DeltaPublishs)==0 && len(deltaModel.DeltaWithdraws)==0, deltaModel:",
			jsonutil.MarshalJson(deltaModel), "   repoPath:", repoPath)
		return rrdpFiles, nil
	}
	belogs.Info("saveRrdpDeltaToRrdpFiles():serial:", deltaModel.Serial,
		"   len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws),
		"   len(rrdpUris):", len(rrdpUris), "    repoPath:", repoPath)
	if len(deltaModel.DeltaWithdraws) > 0 {
		belogs.Info("saveRrdpDeltaToRrdpFiles():len(deltaModel.DeltaWithdraws)>0, deltaModel:", jsonutil.MarshalJson(deltaModel))
	}
	rrdpFiles = make([]RrdpFile, 0)

	// first , del withdraw files
	for i := range deltaModel.DeltaWithdraws {
		uri := deltaModel.DeltaWithdraws[i].Uri
		belogs.Debug("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws, uri:", uri)
		uri = strings.Replace(uri, "../", "/", -1) //fix Path traversal
		belogs.Debug("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws Replace, uri:", uri)
		if v, ok := rrdpUris[uri]; ok {
			belogs.Info("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws in rrdpUris , uri:", uri,
				"    this:", jsonutil.MarshalJson(deltaModel.DeltaWithdraws[i]),
				"    last:", jsonutil.MarshalJson(v))
			continue
		} else {
			rrdpUris[uri] = deltaModel.Serial
		}

		pathFileName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, uri)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws JoinPrefixPathAndUrlFileName fail,uri:", uri, err)
			return nil, err
		}

		// if in this dir, no more files, then del dir
		// will ignore error
		dir, file := osutil.Split(pathFileName)
		if !fileutil.CheckPathNameMaxLength(dir) {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws CheckPathNameMaxLength fail,dir:", dir)
			return nil, errors.New("DeltaWithdraw path name is too long")
		}
		if !fileutil.CheckFileNameMaxLength(file) {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws CheckFileNameMaxLength fail,file:", file)
			return nil, errors.New("DeltaWithdraw file name is too long")
		}
		files, err := ioutil.ReadDir(dir)
		belogs.Info("saveRrdpDeltaToRrdpFiles():DeltaWithdraws will remove pathFileName, uri:", uri,
			"  	pathFileName:", pathFileName, "   dir:", dir,
			"   files:", len(files), "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
			"   err:", err)
		exist, err := osutil.IsExists(pathFileName)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles():DeltaWithdraws IsExists pathFileName fail:", pathFileName,
				"   dir:", dir, "   files:", len(files), "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
				"   err:", err)
			// ignore return
		}
		if exist {
			err = os.Remove(pathFileName)
			if err != nil {
				belogs.Error("saveRrdpDeltaToRrdpFiles():DeltaWithdraws remove pathFileName fail:", pathFileName,
					"   dir:", dir, "   files:", len(files), "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
					"   err:", err)
				// ignore return
			}
		}
		if len(files) == 0 {
			err = os.RemoveAll(dir)
			if err != nil {
				belogs.Error("saveRrdpDeltaToRrdpFiles():DeltaWithdraws RemoveAll dir fail:", pathFileName,
					"   dir:", dir, "   files:", len(files), "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
					"   err:", err)
				// ignore return
			}
		}

		rrdpFile := RrdpFile{
			FilePath:  dir,
			FileName:  file,
			SyncType:  "del",
			SourceUrl: deltaModel.DeltaUrl,
		}
		belogs.Info("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws, del pathFileName:", pathFileName,
			"    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
			"    rrdpFile:", jsonutil.MarshalJson(rrdpFile), "  ok")
		belogs.Debug("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws,  pathFileName:", pathFileName,
			"   dir:", dir, "   rrdpFile:", jsonutil.MarshalJson(rrdpFile),
			"   deltaModel.DeltaUrl:", deltaModel.DeltaUrl, "  ok")

		rrdpFiles = append(rrdpFiles, rrdpFile)
	}
	belogs.Info("saveRrdpDeltaToRrdpFiles():after DeltaWithdraws, len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws),
		"   len(rrdpFiles):", len(rrdpFiles), "   len(rrdpUris):", len(rrdpUris))

	// seconde, save publish files
	for i := range deltaModel.DeltaPublishs {
		uri := deltaModel.DeltaPublishs[i].Uri
		belogs.Debug("saveRrdpDeltaToRrdpFiles(): DeltaPublishs, uri:", uri)
		uri = strings.Replace(uri, "../", "/", -1) //fix Path traversal
		belogs.Debug("saveRrdpDeltaToRrdpFiles(): DeltaPublishs Replace, uri:", uri)
		if v, ok := rrdpUris[uri]; ok {
			belogs.Info("saveRrdpDeltaToRrdpFiles(): DeltaPublishs in rrdpUris, uri:", uri,
				"    this:", jsonutil.MarshalJson(deltaModel.DeltaPublishs[i]),
				"    last:", jsonutil.MarshalJson(v))
			continue
		} else {
			rrdpUris[uri] = deltaModel.Serial
		}

		// get absolute dir /dest/***/***/**.**
		pathFileName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, uri)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): JoinPrefixPathAndUrlFileName fail, uri:",
				uri, "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl, err)
			return nil, err
		}

		// if dir is notexist ,then mkdir
		dir, file := osutil.Split(pathFileName)
		if !fileutil.CheckPathNameMaxLength(dir) {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): Publish CheckPathNameMaxLength fail,dir:", dir)
			return nil, errors.New("Publish path name is too long")
		}
		if !fileutil.CheckFileNameMaxLength(file) {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): Publish CheckFileNameMaxLength fail,file:", file)
			return nil, errors.New("Publish file name is too long")
		}

		isExist, err := osutil.IsExists(dir)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): Publish ReadDir fail:", dir, "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl, err)
			return nil, err
		}
		if !isExist {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				belogs.Error("saveRrdpDeltaToRrdpFiles(): Publish MkdirAll fail:", dir, "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
					err)
				return nil, err
			}
		}

		// decode base65 to bytes
		bytes, err := base64util.DecodeBase64(strings.TrimSpace(deltaModel.DeltaPublishs[i].Base64))
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles():Publish DecodeBase64 fail:",
				"  deltaModel.Serial:", deltaModel.Serial,
				"  deltaModel.DeltaPublishs[i].Uri:", uri,
				"  deltaModel.DeltaPublishs[i].Base64:", deltaModel.DeltaPublishs[i].Base64,
				"  deltaModel.DeltaUrl:", deltaModel.DeltaUrl, err)
			return nil, err
		}

		err = fileutil.WriteBytesToFile(pathFileName, bytes)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles():Publish WriteBytesToFile fail:",
				"  deltaModel.Serial:", deltaModel.Serial,
				"  deltaModel.DeltaPublishs[i].Uri:", uri,
				"  pathFileName:", pathFileName,
				"  deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
				"  len(bytes):", len(bytes),
				err)
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
		belogs.Info("saveRrdpDeltaToRrdpFiles(): Publish, update pathFileName:", pathFileName,
			"  deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
			"  rrdpFile:", jsonutil.MarshalJson(rrdpFile), "  ok")
		belogs.Debug("saveRrdpDeltaToRrdpFiles():Publish update rrdpFile ", jsonutil.MarshalJson(rrdpFile), "  ok")
		rrdpFiles = append(rrdpFiles, rrdpFile)
	}
	belogs.Info("SaveRrdpSnapshotToRrdpFiles(): after all, len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws),
		"   len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(rrdpFiles): ", len(rrdpFiles), "   len(rrdpUris):", len(rrdpUris))
	belogs.Debug("SaveRrdpSnapshotToRrdpFiles(): save rrdpFiles: ", jsonutil.MarshalJson(rrdpFiles))
	return rrdpFiles, nil

}
