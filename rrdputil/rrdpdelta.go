package rrdputil

import (
	"errors"
	"net/http"
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

/*
// deprecated: please use GetRrdpDeltasWithConfig

	func GetRrdpDeltas(notificationModel *NotificationModel, lastSerial uint64) (deltaModels []DeltaModel, err error) {
		belogs.Info("GetRrdpDeltas(): len(notificationModel.Deltas):", len(notificationModel.Deltas), "  lastSerial:", lastSerial)
		return GetRrdpDeltasWithConfig(notificationModel, lastSerial, nil)
	}
*/
func GetRrdpDeltasWithConfig(notificationModel *NotificationModel, lastSerial uint64, httpClientConfig *httpclient.HttpClientConfig) (deltaModels []DeltaModel, err error) {
	start := time.Now()
	if httpClientConfig == nil {
		httpClientConfig = httpclient.CloneGLobalHttpClient()
	}
	belogs.Info("GetRrdpDeltasWithConfig(): len(notificationModel.Deltas):", len(notificationModel.Deltas),
		"  lastSerial:", lastSerial, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))

	var wg sync.WaitGroup
	errorMsgCh := make(chan string, len(notificationModel.Deltas))
	deltaModelCh := make(chan DeltaModel, len(notificationModel.Deltas))
	countCh := make(chan int, runtime.NumCPU()*2)
	// serial need from newest to oldest
	for i := 0; i < len(notificationModel.Deltas); i++ {
		belogs.Debug("GetRrdpDeltasWithConfig(): i:", i, "   notificationModel.Deltas[i].Serial:", notificationModel.Deltas[i].Serial)
		if notificationModel.Deltas[i].Serial <= lastSerial {
			belogs.Debug("GetRrdpDeltasWithConfig():continue, notificationModel.Deltas[i].Serial <= lastSerial:", notificationModel.Deltas[i].Serial, lastSerial)
			continue
		}

		countCh <- 1
		wg.Add(1)
		go getRrdpDeltasImplWithConfig(notificationModel, i, httpClientConfig, deltaModelCh, errorMsgCh, countCh, &wg)
		//
	}
	wg.Wait()
	close(countCh)
	close(errorMsgCh)
	close(deltaModelCh)

	belogs.Debug("GetRrdpDeltasWithConfig():will get errorMsgCh, and deltaModelCh", len(errorMsgCh), len(deltaModelCh))
	// if has error, then return error
	for errorMsg := range errorMsgCh {
		belogs.Error("GetRrdpDeltasWithConfig(): getRrdpDeltasImpl fail:", errorMsg, "   time(s):", time.Since(start))
		return nil, errors.New(errorMsg)
	}
	// get deltaModels, and sort
	deltaModels = make([]DeltaModel, 0, len(notificationModel.MapSerialDeltas))
	for deltaModel := range deltaModelCh {
		deltaModels = append(deltaModels, deltaModel)
	}
	// sort, from newest to oldest
	sort.Sort(DeltaModelsSort(deltaModels))

	belogs.Info("GetRrdpDeltasWithConfig():len(deltaModels):", len(deltaModels),
		"   len(notificationModel.Deltas) :", len(notificationModel.Deltas),
		"   lastSerial:", lastSerial, "   time(s):", time.Since(start))

	return deltaModels, nil
}

func getRrdpDeltasImplWithConfig(notificationModel *NotificationModel, i int, httpClientConfig *httpclient.HttpClientConfig,
	deltaModelCh chan DeltaModel, errorMsgCh chan string,
	countCh chan int, wg *sync.WaitGroup) {
	defer func() {
		<-countCh
		wg.Done()
	}()

	start := time.Now()
	belogs.Debug("getRrdpDeltasImplWithConfig():will notificationModel.Deltas[i].Uri:", i, notificationModel.Deltas[i].Uri)
	deltaModel, err := GetRrdpDeltaWithConfig(notificationModel.Deltas[i].Uri, httpClientConfig)
	if err != nil {
		belogs.Error("getRrdpDeltasImplWithConfig(): GetRrdpDelta fail, delta.Uri :", i,
			notificationModel.Deltas[i].Uri, err, "   time(s):", time.Since(start))
		errorMsgCh <- "get delta " + notificationModel.Deltas[i].Uri + " fail, error is " + err.Error()
		return
	}
	belogs.Debug("getRrdpDeltasImplWithConfig():ok notificationModel.Deltas[i].Uri:", i, notificationModel.Deltas[i].Uri,
		"   time(s):", time.Since(start))

	err = CheckRrdpDelta(&deltaModel, notificationModel)
	if err != nil {
		belogs.Error("getRrdpDeltasImplWithConfig(): CheckRrdpDelta fail, delta.Uri :", i,
			notificationModel.Deltas[i].Uri, err, "   time(s):", time.Since(start))
		errorMsgCh <- "check delta " + notificationModel.Deltas[i].Uri + " fail, error is " + err.Error()
		return
	}
	belogs.Debug("getRrdpDeltasImplWithConfig(): delta.Uri:", notificationModel.Deltas[i].Uri,
		"   len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws),
		"   time(s):", time.Since(start))
	deltaModelCh <- deltaModel
	return
}

// deprecated: please use GetRrdpDeltaWithConfig
func GetRrdpDelta(deltaUrl string) (deltaModel DeltaModel, err error) {
	belogs.Debug("GetRrdpDelta(): deltaUrl:", deltaUrl)
	return GetRrdpDeltaWithConfig(deltaUrl, nil)
}
func GetRrdpDeltaWithConfig(deltaUrl string, httpClientConfig *httpclient.HttpClientConfig) (deltaModel DeltaModel, err error) {

	start := time.Now()
	if httpClientConfig == nil {
		httpClientConfig = httpclient.CloneGLobalHttpClient()
	}
	// get delta.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/43230/delta.xml"
	belogs.Debug("GetRrdpDeltaWithConfig(): deltaUrl:", deltaUrl, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	deltaModel, err = getRrdpDeltaImplWithConfig(deltaUrl, httpClientConfig)
	if err != nil {
		belogs.Error("GetRrdpDeltaWithConfig():getRrdpDeltaImpl fail:", deltaUrl, err)
		return deltaModel, err
	}

	belogs.Info("GetRrdpDeltaWithConfig(): deltaUrl ok:", deltaUrl, "  time(s):", time.Since(start))
	return deltaModel, nil
}

func getRrdpDeltaImplWithConfig(deltaUrl string, httpClientConfig *httpclient.HttpClientConfig) (deltaModel DeltaModel, err error) {

	// get delta.xml
	// "https://rrdp.apnic.net/4ea5d894-c6fc-4892-8494-cfd580a414e3/43230/delta.xml"
	belogs.Debug("getRrdpDeltaImplWithConfig(): deltaUrl:", deltaUrl, "  httpClientConfig:", jsonutil.MarshalJson(httpClientConfig))
	deltaUrl = strings.TrimSpace(deltaUrl)
	start := time.Now()
	resp, body, err := httpclient.GetHttpsVerifyWithConfig(deltaUrl, httpClientConfig)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			belogs.Error("getRrdpDeltaImplWithConfig(): GetHttpsVerifyWithConfig deltaUrl, is not StatusOK:", deltaUrl,
				"   statusCode:", httpclient.GetStatusCode(resp), "    body:", body)
			return deltaModel, errors.New("http status code of " + deltaUrl + " is " + resp.Status)
		} else {
			belogs.Debug("getRrdpDeltaImplWithConfig(): GetHttpsVerifyWithConfig deltaUrl ok:", deltaUrl,
				"   statusCode:", httpclient.GetStatusCode(resp),
				"   ipAddrs:", netutil.LookupIpByUrl(deltaUrl),
				"   len(body):", len(body), "  time(s):", time.Since(start))
		}
	} else {
		belogs.Debug("getRrdpDeltaImplWithConfig(): GetHttpsVerifyWithConfig deltaUrl fail, will use curl again:", deltaUrl, "   ipAddrs:", netutil.LookupIpByUrl(deltaUrl),
			"   resp:", resp, "    len(body):", len(body), "  time(s):", time.Since(start), err)

		// then try using curl
		start = time.Now()
		httpClientConfig.IpType = "ipv4"
		body, err = httpclient.GetByCurlWithConfig(deltaUrl, httpClientConfig)
		if err != nil {
			belogs.Debug("getRrdpDeltaImplWithConfig(): GetByCurlWithConfig deltaUrl fail:", deltaUrl, "   resp:", resp,
				"   ipAddrs:", netutil.LookupIpByUrl(deltaUrl),
				"   len(body):", len(body), "  time(s):", time.Since(start), err)
			// then try again using curl, using all
			start = time.Now()
			httpClientConfig.IpType = "all"
			body, err = httpclient.GetByCurlWithConfig(deltaUrl, httpClientConfig)
			if err != nil {
				belogs.Error("getRrdpDeltaImplWithConfig(): GetByCurlWithConfig deltaUrl, iptype is all, fail:", deltaUrl,
					"   ipAddrs:", netutil.LookupIpByUrl(deltaUrl), "   resp:", resp,
					"   len(body):", len(body), "  time(s):", time.Since(start), err)
				return deltaModel, errors.New("http error of " + deltaUrl + " is " + err.Error())
			}
			belogs.Debug("getRrdpDeltaImplWithConfig(): GetByCurlWithConfig deltaUrl, iptype is all, ok", deltaUrl, "    len(body):", len(body),
				"  time(s):", time.Since(start))
		} else {
			belogs.Debug("getRrdpDeltaImplWithConfig(): GetByCurlWithConfig deltaUrl, iptype is ipv4, ok", deltaUrl, "    len(body):", len(body), "  time(s):", time.Since(start))
		}
	}

	// check if body is xml file
	belogs.Debug("getRrdpDeltaImplWithConfig(): get body, deltaUrl:", deltaUrl, " len(body):", len(body))
	if !strings.Contains(body, `<delta`) {
		belogs.Error("getRrdpDeltaImplWithConfig(): body is not xml file:", deltaUrl, "   resp:",
			resp, "    len(body):", len(body), "       body:", body, "  time(s):", time.Since(start), err)
		return deltaModel, errors.New("body of " + deltaUrl + " is not xml")
	}

	err = xmlutil.UnmarshalXml(body, &deltaModel)
	if err != nil {
		belogs.Error("getRrdpDeltaImplWithConfig(): UnmarshalXml fail:", deltaUrl, "    body:", body, err)
		return deltaModel, err
	}

	deltaModel.Hash = hashutil.Sha256([]byte(body))
	belogs.Debug("getRrdpDeltaImplWithConfig(): len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws))
	for i := range deltaModel.DeltaPublishs {
		uri := strings.Replace(deltaModel.DeltaPublishs[i].Uri, "../", "/", -1) //fix Path traversal
		deltaModel.DeltaPublishs[i].Uri = uri
		deltaModel.DeltaPublishs[i].Base64 = stringutil.TrimSpaceAndNewLine(deltaModel.DeltaPublishs[i].Base64)
	}
	for i := range deltaModel.DeltaWithdraws {
		uri := strings.Replace(deltaModel.DeltaWithdraws[i].Uri, "../", "/", -1) //fix Path traversal
		deltaModel.DeltaWithdraws[i].Uri = uri
	}
	deltaModel.DeltaUrl = deltaUrl
	belogs.Info("getRrdpDeltaImplWithConfig(): get from deltaUrl ok", deltaUrl,
		"   len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws), "  time(s):", time.Since(start))
	return deltaModel, nil
}

func CheckRrdpDeltaValue(deltaModel *DeltaModel) (err error) {
	if deltaModel.Version != "1" {
		belogs.Error("CheckRrdpDeltaValue(): deltaModel.Version != 1, current delta version is outdated, url is " + deltaModel.DeltaUrl)
		return errors.New("current delta version is outdated, url is " + deltaModel.DeltaUrl)
	}
	if len(deltaModel.SessionId) == 0 {
		belogs.Error("CheckRrdpDeltaValue(): len(deltaModel.SessionId) == 0")
		return errors.New("delta session_id is error, session_id is empty  ")
	}
	return nil
}

func CheckRrdpDelta(deltaModel *DeltaModel, notificationModel *NotificationModel) (err error) {
	err = CheckRrdpDeltaValue(deltaModel)
	if err != nil {
		belogs.Error("CheckRrdpDelta(): CheckRrdpDeltaValue fail:", err)
		return err
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
		if v, ok := rrdpUris[uri]; ok {
			belogs.Info("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws in rrdpUris , uri:", uri,
				"    this:", jsonutil.MarshalJson(deltaModel.DeltaWithdraws[i]),
				"    last:", jsonutil.MarshalJson(v))
			continue
		} else {
			rrdpUris[uri] = deltaModel.Serial
		}

		filePathName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, uri)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws JoinPrefixPathAndUrlFileName fail,uri:", uri, err)
			return nil, err
		}

		// if in this dir, no more files, then del dir
		// will ignore error
		dir, file := osutil.Split(filePathName)
		if !fileutil.CheckPathNameMaxLength(dir) {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws CheckPathNameMaxLength fail,dir:", dir)
			return nil, errors.New("DeltaWithdraw path name is too long")
		}
		if !fileutil.CheckFileNameMaxLength(file) {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws CheckFileNameMaxLength fail,file:", file)
			return nil, errors.New("DeltaWithdraw file name is too long")
		}
		files, err := os.ReadDir(dir)
		belogs.Info("saveRrdpDeltaToRrdpFiles():DeltaWithdraws will remove filePathName, uri:", uri,
			"  	filePathName:", filePathName, "   dir:", dir,
			"   files:", len(files), "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
			"   err:", err)
		exist, err := osutil.IsExists(filePathName)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles():DeltaWithdraws IsExists filePathName fail:", filePathName,
				"   dir:", dir, "   files:", len(files), "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
				"   err:", err)
			// ignore return
		}
		if exist {
			err = os.Remove(filePathName)
			if err != nil {
				belogs.Error("saveRrdpDeltaToRrdpFiles():DeltaWithdraws remove filePathName fail:", filePathName,
					"   dir:", dir, "   files:", len(files), "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
					"   err:", err)
				// ignore return
			}
		}
		if len(files) == 0 {
			err = os.RemoveAll(dir)
			if err != nil {
				belogs.Error("saveRrdpDeltaToRrdpFiles():DeltaWithdraws RemoveAll dir fail:", filePathName,
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
		belogs.Info("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws, del filePathName:", filePathName,
			"    deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
			"    rrdpFile:", jsonutil.MarshalJson(rrdpFile), "  ok")
		belogs.Debug("saveRrdpDeltaToRrdpFiles(): DeltaWithdraws,  filePathName:", filePathName,
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
		if v, ok := rrdpUris[uri]; ok {
			belogs.Info("saveRrdpDeltaToRrdpFiles(): DeltaPublishs in rrdpUris, uri:", uri,
				"    this:", jsonutil.MarshalJson(deltaModel.DeltaPublishs[i]),
				"    last:", jsonutil.MarshalJson(v))
			continue
		} else {
			rrdpUris[uri] = deltaModel.Serial
		}

		// get absolute dir /dest/***/***/**.**
		filePathName, err := urlutil.JoinPrefixPathAndUrlFileName(repoPath, uri)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles(): JoinPrefixPathAndUrlFileName fail, uri:",
				uri, "    deltaModel.DeltaUrl:", deltaModel.DeltaUrl, err)
			return nil, err
		}

		// if dir is notexist ,then mkdir
		dir, file := osutil.Split(filePathName)
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

		err = fileutil.WriteBytesToFile(filePathName, bytes)
		if err != nil {
			belogs.Error("saveRrdpDeltaToRrdpFiles():Publish WriteBytesToFile fail:",
				"  deltaModel.Serial:", deltaModel.Serial,
				"  deltaModel.DeltaPublishs[i].Uri:", uri,
				"  filePathName:", filePathName,
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
		belogs.Info("saveRrdpDeltaToRrdpFiles(): Publish, update filePathName:", filePathName,
			"  deltaModel.DeltaUrl:", deltaModel.DeltaUrl,
			"  rrdpFile:", jsonutil.MarshalJson(rrdpFile), "  ok")
		belogs.Debug("saveRrdpDeltaToRrdpFiles():Publish update rrdpFile ", jsonutil.MarshalJson(rrdpFile), "  ok")
		rrdpFiles = append(rrdpFiles, rrdpFile)
	}
	belogs.Info("saveRrdpDeltaToRrdpFiles(): after all, len(deltaModel.DeltaWithdraws):", len(deltaModel.DeltaWithdraws),
		"   len(deltaModel.DeltaPublishs):", len(deltaModel.DeltaPublishs),
		"   len(rrdpFiles): ", len(rrdpFiles), "   len(rrdpUris):", len(rrdpUris))
	belogs.Debug("saveRrdpDeltaToRrdpFiles(): save rrdpFiles: ", jsonutil.MarshalJson(rrdpFiles))
	return rrdpFiles, nil

}
