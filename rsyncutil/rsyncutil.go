package rsyncutil

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/transportutil"
	"github.com/cpusoft/goutil/urlutil"
)

// rsync type

var rsyncClientConfig = NewRsyncClientConfig()

type RsyncRecord struct {
	Id           uint64        `json:"id"`
	StartTime    time.Time     `json:"startTime"`
	EndTime      time.Time     `json:"endTime"`
	Style        string        `json:"style"`
	RsyncResults []RsyncResult `json:"rsyncResults"`
}

type RsyncResult struct {
	Id       uint64 `json:"id"`
	RsyncId  uint64 `json:"rsyncId"`
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	FileType string `json:"fileType"`
	//RSYNC_TYPE_***
	RsyncType string    `json:"rsyncType"`
	RsyncUrl  string    `json:"rsyncUrl"`
	IsDir     bool      `json:"isDir"`
	SyncTime  time.Time `json:"syncTime"`
}

func Rsync(rsyncUrl, destPath string) (rsyncResults []RsyncResult, err error) {
	belogs.Debug("Rsync():rsyncUrl:", rsyncUrl, " destPath:", destPath)

	rsyncDestPath, output, err := RsyncToStdout(rsyncUrl, destPath)
	if err != nil {
		belogs.Error("Rsync():RsyncToStdout fail, rsyncUrl:", rsyncUrl, "   err:", err)
		return nil, err
	}

	rsyncResults, err = ParseStdoutToRsyncResults(rsyncUrl, rsyncDestPath, output)
	if err != nil {
		belogs.Error("Rsync():ParseStdoutToRsyncResults fail, rsyncUrl:", rsyncUrl, "   rsyncDestPath:", rsyncDestPath, "   err:", err)
		return nil, err
	}

	belogs.Debug("Rsync():before AddCerToRsyncResults, rsyncDestPath:", rsyncDestPath, "   len(rsyncResults)", len(rsyncResults))
	err = AddCerToRsyncResults(rsyncDestPath, rsyncResults)
	if err != nil {
		belogs.Error("Rsync():AddCerToRsyncResults fail, rsyncUrl:", rsyncUrl, "   rsyncDestPath:", rsyncDestPath, "   err:", err)
		return nil, err
	}
	belogs.Debug("Rsync():after AddCerToRsyncResults, rsyncDestPath:", rsyncDestPath, "   len(rsyncResults)", len(rsyncResults))
	return rsyncResults, err
}

// set rsync url and local dest path , then will call rsync
// will rsync every file but no output
// if success, the len(output) will be zero
func RsyncQuiet(rsyncUrl string, destPath string) (rsyncDestPath string, output []byte, err error) {
	belogs.Debug("RsyncQuiet():rsyncUrl:", rsyncUrl, " destPath:", destPath)
	defer func(rsyncUrl string) {
		if err := recover(); err != nil {
			errStack := string(debug.Stack())
			belogs.Error("RsyncQuiet(): recover from panic, rsyncUrl is :", rsyncUrl,
				" debug.Stack():", errStack, "  err is :", err)

		}
	}(rsyncUrl)

	// get host+path by url
	hostAndPath, err := urlutil.HostAndPath(rsyncUrl)
	belogs.Debug("RsyncQuiet():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath)
	if err != nil {
		belogs.Error("RsyncQuiet():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath, " err:", err)
		return "", output, err
	}

	// mkdirAll path
	rsyncDestPath = osutil.JoinPathFile(destPath, hostAndPath)
	belogs.Debug("RsyncQuiet():rsyncDestPath:", rsyncDestPath)
	err = os.MkdirAll(rsyncDestPath, os.ModePerm)
	if err != nil {
		belogs.Error("RsyncQuiet():MkdirAll:", rsyncDestPath, " err:", err)
		return "", output, err
	}

	// call rsync
	//rsync -Lirzts --del --timeout=5 --contimeout=5 --no-motd  -4 rsync://rpki.afrinic.net/repository/afrinic/  /tmp/rpki.afrinic.net/repository/afrinic/
	//-L  --copy-links            transform symlink into referent file/dir
	//-r  --recursive             recurse into directories
	//-z  --compress              compress file data during the transfer
	//-t  --times                 preserve modification times
	//-s  --protect-args          no space-splitting; only wildcard special-chars
	//--del                   an alias for --delete-during
	//--delete-during         receiver deletes during the transfer
	//-4  --ipv4                  prefer IPv4
	//--timeout=SECONDS       set I/O timeout in seconds
	//--no-motd               suppress daemon-mode MOTD (see manpage caveat)
	//--contimeout=SECONDS    set daemon connection timeout in seconds
	belogs.Debug("RsyncQuiet(): Command: rsync", "-Lirzts", "--del", "--timeout="+rsyncClientConfig.Timeout, "--contimeout="+rsyncClientConfig.ConTimeout,
		"--no-motd", "-4", rsyncUrl, rsyncDestPath)
	cmd := exec.Command("rsync", "-Lrzts", "--del", "--timeout="+rsyncClientConfig.Timeout, "--contimeout="+rsyncClientConfig.ConTimeout,
		"--no-motd", "-4", rsyncUrl, rsyncDestPath)
	// if success, the len(output) will be zero
	output, err = cmd.CombinedOutput()
	if err != nil {
		belogs.Error("RsyncQuiet(): exec.Command fail, rsyncUrl is :", rsyncUrl, "   output is ", string(output), " err is :", err)
		// some err detail in output
		err = errors.New(string(output) + ", " + err.Error())
		return "", output, err
	}
	belogs.Debug("RsyncQuiet(): rsyncDestPath:", rsyncDestPath, "  output:", string(output))
	return rsyncDestPath, output, nil
}

// use telnet test connect
func RsyncTestConnect(rsyncUrl string) (err error) {
	start := time.Now()
	belogs.Debug("RsyncTestConnect():rsyncUrl:", rsyncUrl)
	host, port, err := urlutil.HostAndPort(rsyncUrl)
	if err != nil {
		belogs.Error("RsyncTestConnect(): HostAndPort fail, rsyncUrl:", rsyncUrl, err, "  time(s):", time.Since(start))
		return errors.New("rsync error of " + rsyncUrl + " is " + err.Error())
	}
	if len(port) == 0 {
		port = "873"
	}

	belogs.Debug("RsyncTestConnect():rsyncUrl:", rsyncUrl,
		"   host:", host, "   port:", port)
	err = transportutil.TestTcpConnection(host, port)
	if err != nil {
		belogs.Error("RsyncTestConnect(): TestTcpConnection fail:", rsyncUrl, err, "  time(s):", time.Since(start))
		return errors.New("rsync error of " + rsyncUrl + " is " + err.Error())
	}
	belogs.Debug("RsyncTestConnect(): ok, rsyncUrl:", rsyncUrl, "  time(s):", time.Since(start))
	return nil
}

// set rsync url and local dest path , then will call rsync
// will get all stdout to get every file rsync state
// and output will save to logPath
func RsyncToLogFile(rsyncUrl string, destPath string, logPath string) (rsyncDestPath, rsyncLogFile string, err error) {
	belogs.Debug("RsyncToLogFile():rsyncUrl:", rsyncUrl, " destPath:", destPath, " logPath:", logPath)

	// get host+path by url
	hostAndPath, err := urlutil.HostAndPath(rsyncUrl)
	belogs.Debug("RsyncToLogFile():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath)
	if err != nil {
		belogs.Error("RsyncToLogFile():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath, " err:", err)
		return "", "", err
	}

	// mkdirAll path
	rsyncDestPath = osutil.JoinPathFile(destPath, hostAndPath)
	belogs.Debug("RsyncToLogFile():rsyncDestPath:", rsyncDestPath)
	err = os.MkdirAll(rsyncDestPath, os.ModePerm)
	if err != nil {
		belogs.Error("RsyncToLogFile():MkdirAll rsyncDestPath:", rsyncDestPath, " err:", err)
		return "", "", err
	}

	// mkdirAll path
	rsyncLogPath := osutil.JoinPathFile(logPath, hostAndPath)
	belogs.Debug("RsyncToLogFile():rsyncLogPath:", rsyncLogPath)
	err = os.MkdirAll(rsyncLogPath, os.ModePerm)
	if err != nil {
		belogs.Error("RsyncToLogFile():MkdirAll rsyncLogPath:", rsyncLogPath, " err:", err)
		return "", "", err
	}
	rsyncLogFile = osutil.JoinPathFile(rsyncLogPath, RSYNC_LOG_FILE_NAME)
	belogs.Debug("RsyncToLogFile():rsyncLogFile:", rsyncLogFile)

	// call rsync
	//rsync -Lirzts --del --timeout=5 --contimeout=5 --no-motd  -4 rsync://rpki.afrinic.net/repository/afrinic/  /tmp/rpki.afrinic.net/repository/afrinic/
	belogs.Debug("RsyncToLogFile(): Command: rsync", "-Lirzts", "--del", "--no-motd", "-4", "--log-file=\""+rsyncLogFile+"\"", rsyncUrl, rsyncDestPath)
	cmd := exec.Command("rsync", "-Lirzts", "--del", "--no-motd", "-4", "--log-file=\""+rsyncLogFile+"\"", rsyncUrl, rsyncDestPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("RsyncToLogFile(): exec.Command fail, rsyncUrl is :", rsyncUrl, "   output is ", string(output), " err is :", err)
		err = errors.New(string(output) + ", " + err.Error())
		return "", "", err
	}
	belogs.Debug("RsyncToLogFile(): rsyncDestPath:", rsyncDestPath, "  rsyncLogFile:", rsyncLogFile)
	return rsyncDestPath, rsyncLogFile, nil

}

// set rsync url and local dest path , then will call rsync
// will get all stdout to get every file rsync state
// and output will convert to rsyncResult
func RsyncToStdout(rsyncUrl string, destPath string) (rsyncDestPath string, output []byte, err error) {
	belogs.Debug("RsyncToStdout():rsyncUrl:", rsyncUrl, " destPath:", destPath)

	//
	output = make([]byte, 0)

	// get host+path by url
	hostAndPath, err := urlutil.HostAndPath(rsyncUrl)
	belogs.Debug("RsyncToStdout():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath)
	if err != nil {
		belogs.Error("RsyncToStdout():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath, " err:", err)
		return "", output, err
	}

	// mkdirAll path
	rsyncDestPath = osutil.JoinPathFile(destPath, hostAndPath)
	belogs.Debug("RsyncToStdout():rsyncDestPath:", rsyncDestPath)
	err = os.MkdirAll(rsyncDestPath, os.ModePerm)
	if err != nil {
		belogs.Error("RsyncToStdout():MkdirAll:", rsyncDestPath, " err:", err)
		return "", output, err
	}

	// call rsync
	//rsync -Lirzts --del --timeout=5 --contimeout=5 --no-motd  -4 rsync://rpki.afrinic.net/repository/afrinic/  /tmp/rpki.afrinic.net/repository/afrinic/
	//-L  --copy-links            transform symlink into referent file/dir
	//-i  --itemize-changes       output a change-summary for all updates
	//-r  --recursive             recurse into directories
	//-z  --compress              compress file data during the transfer
	//-t  --times                 preserve modification times
	//-s  --protect-args          no space-splitting; only wildcard special-chars
	//--del                   an alias for --delete-during
	//--delete-during         receiver deletes during the transfer
	//-4  --ipv4                  prefer IPv4
	//--timeout=SECONDS       set I/O timeout in seconds
	//--no-motd               suppress daemon-mode MOTD (see manpage caveat)
	belogs.Debug("RsyncToStdout(): Command: rsync", "-Lirzts", "--del", "--timeout=600", "--no-motd", "-4", rsyncUrl, rsyncDestPath)
	cmd := exec.Command("rsync", "-Lirzts", "--del", "--timeout=6000", "--no-motd", "-4", rsyncUrl, rsyncDestPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		belogs.Alert("RsyncToStdout(): exec.Command fail, rsyncUrl is :", rsyncUrl, "   output is ", string(output), " err is :", err)
		// some err detail in output
		err = errors.New(string(output) + ", " + err.Error())
		return "", output, err
	}
	belogs.Debug("RsyncToStdout(): rsyncDestPath:", rsyncDestPath, "  output:", string(output))
	return rsyncDestPath, output, nil
}

// get results from logfile
func ParseLogfileToRsyncResults(rsyncUrl, rsyncDestPath, rsyncLogFile string) (rsyncResults []RsyncResult, err error) {
	// if some changed ,so length of output is > 0
	rsyncResults = make([]RsyncResult, 0)
	lines, err := fileutil.ReadFileToLines(rsyncLogFile)
	if err != nil {
		belogs.Error("ParseLogfileToRsyncResults(): ReadFileToLines: err: ", rsyncLogFile, err)
		return rsyncResults, err
	}
	if len(lines) == 0 {
		return rsyncResults, nil
	}
	//2019/10/22 18:13:36 [12777] receiving file list
	//2019/10/22 18:11:02 [12313] >f+++++++++ B527EF581D6611E2BB468F7C72FD1FF2/zz57dCpMUuD49_jnQF-yu7z0yqM.cer
	//2019/10/22 18:13:36 [12777] sent 43 bytes  received 66591 bytes  total size 2844287
	results := make([]string, 0)
	for _, one := range lines {
		if strings.Index(one, "] receiving file list") > 0 ||
			strings.Index(one, "] sent") > 0 {
			continue
		}
		pos := strings.Index(one, "]")
		one = string(one[pos+2:])
		results = append(results, one)
	}

	belogs.Debug("ParseLogfileToRsyncResults(): len(results):", len(results))
	return parseToRsyncResults(rsyncUrl, rsyncDestPath, results)
}

// get results from stdout
func ParseStdoutToRsyncResults(rsyncUrl string, rsyncDestPath string, output []byte) (rsyncResults []RsyncResult, err error) {
	// if some changed ,so length of output is > 0
	if len(output) == 0 {
		return rsyncResults, nil
	}
	result := string(output)
	results := strings.Split(result, osutil.GetNewLineSep())
	belogs.Debug("ParseStdoutToRsyncResults(): len(results):", len(results))
	return parseToRsyncResults(rsyncUrl, rsyncDestPath, results)

}

func parseToRsyncResults(rsyncUrl string, rsyncDestPath string, results []string) (rsyncResults []RsyncResult, err error) {
	belogs.Debug("parseToRsyncResults():destPath:", rsyncDestPath, " result:", results)
	rsyncResults = make([]RsyncResult, 0)
	if !strings.HasSuffix(rsyncDestPath, string(os.PathSeparator)) {
		rsyncDestPath = rsyncDestPath + string(os.PathSeparator)
	}
	for _, result := range results {
		if len(result) <= RSYNC_LOG_PREFIX {
			continue
		}
		result = strings.Replace(result, "\n", "", -1)
		result = strings.Replace(result, "\r", "", -1)
		first := result[0]
		second := result[1]
		rsyncResult := RsyncResult{}
		/*
			https://stackoverflow.com/questions/40612505/format-of-rsync-logfile
					* : delete
					< : ignore
					> :
					    ++++: add
					    ....: update
					L:  link  , ignore
					h:  link  , ignore
					c:
						d: directory,    cd+++++++++
						L: link ,cL  , ignore
					.: attribute change ,ignore
		*/

		switch first {
		case '*':
			rsyncResult.RsyncType = RSYNC_TYPE_DEL
			rsyncResult.FilePath, rsyncResult.FileName =
				osutil.GetFilePathAndFileName(rsyncDestPath + string(result[RSYNC_LOG_PREFIX:]))
			rsyncResult.FileType = strings.Replace(path.Ext(rsyncResult.FileName), ".", "", -1)
		case '>':
			if strings.Contains(result, "++++") {
				rsyncResult.RsyncType = RSYNC_TYPE_ADD
			} else {
				rsyncResult.RsyncType = RSYNC_TYPE_UPDATE
			}
			rsyncResult.FilePath, rsyncResult.FileName =
				osutil.GetFilePathAndFileName(rsyncDestPath + string(result[RSYNC_LOG_PREFIX:]))
			rsyncResult.FileType = strings.Replace(path.Ext(rsyncResult.FileName), ".", "", -1)
		case 'c':
			switch second {
			case 'd':
				rsyncResult.RsyncType = RSYNC_TYPE_MKDIR
				rsyncResult.FilePath = rsyncDestPath + string(result[RSYNC_LOG_PREFIX:])

			default:
				rsyncResult.RsyncType = RSYNC_TYPE_IGNORE
			}
		case '<':
			fallthrough
		case 'L':
			fallthrough
		case 'h':
			fallthrough
		case '.':
			fallthrough
		default:
			rsyncResult.RsyncType = RSYNC_TYPE_IGNORE
		}
		rsyncResult.SyncTime = time.Now()
		rsyncResult.IsDir, _ = osutil.IsDir(rsyncResult.FilePath + rsyncResult.FileName)
		belogs.Info("parseToRsyncResults():rsyncResult:", jsonutil.MarshalJson(rsyncResult))
		if err != nil {
			belogs.Error("parseToRsyncResults(): parseRsyncResult: err: ", err, ": "+result)
			return rsyncResults, err
		}
		if rsyncResult.RsyncType == RSYNC_TYPE_IGNORE ||
			len(rsyncResult.FileName) == 0 || len(rsyncResult.FilePath) == 0 {
			belogs.Debug("parseToRsyncResults():ignore parseRsyncResult: ", rsyncResult)
			continue
		}
		rsyncResult.RsyncUrl = rsyncUrl
		rsyncResults = append(rsyncResults, rsyncResult)
	}
	belogs.Debug("parseToRsyncResults():rsyncResults: ", jsonutil.MarshalJson(rsyncResults))
	return rsyncResults, nil

}

//  need read all current existed cer file, to just to trigger sub ca repo sync
func AddCerToRsyncResults(rsyncDestPath string, rsyncResults []RsyncResult) (err error) {

	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	files, err := osutil.GetFilesInDir(rsyncDestPath, m)
	belogs.Debug("addCerToRsyncResults():GetFilesInDir, files:", files, err)
	if err != nil {
		belogs.Debug("addCerToRsyncResults():GetFilesInDir, files:", files, err)
		return err
	}
	for _, file := range files {
		belogs.Debug("addCerToRsyncResults():file:", file)
		found := false
		// if this file had synced, should not repeated add
		for _, rsyncResult := range rsyncResults {
			if file == rsyncResult.FileName {
				found = true
				break
			}

		}
		belogs.Debug("addCerToRsyncResults():GetFilesInDir,found:", found, "   file:", file)
		if !found {
			rsyncResult := RsyncResult{}
			rsyncResult.RsyncType = RSYNC_TYPE_JUST_SYNC
			rsyncResult.FilePath = rsyncDestPath
			rsyncResult.FileName = file
			rsyncResult.FileType = "cer"
			rsyncResult.SyncTime = time.Now()
			rsyncResult.IsDir, _ = osutil.IsDir(file)
			rsyncResults = append(rsyncResults, rsyncResult)
			belogs.Info("addCerToRsyncResults(): manual add rsyncResult:", jsonutil.MarshalJson(rsyncResult))
		}
	}
	return nil

}

func GetFilesHashFromDisk(destPath string) (files map[string]RsyncFileHash, err error) {
	start := time.Now()
	path := destPath
	if !strings.HasSuffix(destPath, osutil.GetPathSeparator()) {
		path = destPath + osutil.GetPathSeparator()
	}
	belogs.Debug("GetFilesHashFromDisk():destPath:", destPath, "  path:", path)

	m := make(map[string]string, 4)
	m[".cer"] = ".cer"
	m[".crl"] = ".crl"
	m[".roa"] = ".roa"
	m[".mft"] = ".mft"

	fileStats, err := osutil.GetAllFileStatsBySuffixs(path, m)
	if err != nil {
		belogs.Error("GetFilesHashFromDisk(): GetAllFileStatsBySuffixs fail:", destPath, err)
		return nil, err
	}
	belogs.Debug("GetFilesHashFromDisk(): len(fileStats):", len(fileStats), "    fileStats:", jsonutil.MarshalJson(fileStats))
	files = make(map[string]RsyncFileHash, len(fileStats))
	for i := range fileStats {
		fileHash := RsyncFileHash{}
		fileHash.FileHash = fileStats[i].Hash256
		fileHash.FileName = fileStats[i].FileName
		fileHash.FilePath = fileStats[i].FilePath
		fileHash.FileType = strings.Replace(osutil.Ext(fileStats[i].FileName), ".", "", -1) //remove dot, should be cer/crl/roa/mft
		files[osutil.JoinPathFile(fileStats[i].FilePath, fileStats[i].FileName)] = fileHash
	}

	belogs.Info("GetFilesHashFromDisk(): len(files):", len(files), "  time(s):", time.Since(start))
	return files, nil

}

// db is old, disk is new
func DiffFiles(filesFromDb, filesFromDisk map[string]RsyncFileHash) (addFiles,
	delFiles, updateFiles, noChangeFiles map[string]RsyncFileHash, err error) {

	start := time.Now()
	// if db is empty, so all filesFromDisk is add
	if len(filesFromDb) == 0 {
		return filesFromDisk, nil, nil, nil, nil
	}

	// if disk is empty, so all filesFromDb is del
	if len(filesFromDisk) == 0 {
		return nil, filesFromDb, nil, nil, nil
	}

	// for db, check. add/update/nochange from disk, del from db
	addFiles = make(map[string]RsyncFileHash, len(filesFromDb))
	delFiles = make(map[string]RsyncFileHash, len(filesFromDb))
	updateFiles = make(map[string]RsyncFileHash, len(filesFromDb))
	noChangeFiles = make(map[string]RsyncFileHash, len(filesFromDb))

	// for db, check disk
	for keyDb, valueDb := range filesFromDb {
		// if found in disk,
		if valueDisk, ok := filesFromDisk[keyDb]; ok {
			// if hash is equal, then save to noChangeFiles, else save to updateFiles
			// and db.jsonall should save as lasjsonall
			if valueDb.FileHash == valueDisk.FileHash {
				valueDisk.LastJsonAll = valueDb.LastJsonAll
				noChangeFiles[keyDb] = valueDisk

			} else {
				valueDisk.LastJsonAll = valueDb.LastJsonAll
				updateFiles[keyDb] = valueDisk
			}
			//have found in disk, then del it in disk map, so remain in disk will be add
			delete(filesFromDisk, keyDb)
		} else {

			// if not found in disk ,then is del, so save to delFiles, and value is db
			delFiles[keyDb] = valueDb
		}
	}
	addFiles = filesFromDisk
	belogs.Debug("DiffFiles(): len(addFiles):", len(addFiles), jsonutil.MarshalJson(addFiles))
	belogs.Debug("DiffFiles(): len(delFiles):", len(delFiles), jsonutil.MarshalJson(delFiles))
	belogs.Debug("DiffFiles(): len(updateFiles):", len(updateFiles), jsonutil.MarshalJson(updateFiles))
	belogs.Debug("DiffFiles(): len(noChangeFiles):", len(noChangeFiles), jsonutil.MarshalJson(noChangeFiles))
	belogs.Debug("DiffFiles(): time(s):", time.Since(start))
	belogs.Info("DiffFiles(): len(addFiles):", len(addFiles), "  len(delFiles):", len(delFiles),
		"  len(updateFiles):", len(updateFiles), "  len(noChangeFiles):", len(noChangeFiles), "  time(s):", time.Since(start))

	return addFiles, delFiles, updateFiles, noChangeFiles, nil

}

func NewRsyncClientConfig() *RsyncClientConfig {
	rsyncClientConfig := new(RsyncClientConfig)
	rsyncClientConfig.Timeout = RSYNC_TIMEOUT_SEC
	rsyncClientConfig.ConTimeout = RSYNC_CONTIMEOUT_SEC
	return rsyncClientConfig
}

// seconds
func SetTimeout(timeoutSec uint64) {
	if timeoutSec > 0 {
		rsyncClientConfig.Timeout = convert.ToString(timeoutSec)
	}
}
func SetConTimeout(conTimeoutSec uint64) {
	if conTimeoutSec > 0 {
		rsyncClientConfig.ConTimeout = convert.ToString(conTimeoutSec)
	}
}
func ResetAllTimeout() {
	rsyncClientConfig.Timeout = RSYNC_TIMEOUT_SEC
	rsyncClientConfig.ConTimeout = RSYNC_CONTIMEOUT_SEC
}
