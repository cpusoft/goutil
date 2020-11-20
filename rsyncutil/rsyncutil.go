package rsyncutil

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"runtime/debug"
	"strings"
	"time"

	belogs "github.com/astaxie/beego/logs"
	fileutil "github.com/cpusoft/goutil/fileutil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	urlutil "github.com/cpusoft/goutil/urlutil"
)

// rsync type
const (
	RSYNC_TYPE_ADD       = "add"
	RSYNC_TYPE_DEL       = "del"
	RSYNC_TYPE_UPDATE    = "update"
	RSYNC_TYPE_MKDIR     = "mkdir"
	RSYNC_TYPE_IGNORE    = "ignore"
	RSYNC_TYPE_JUST_SYNC = "justsync" //The file itself is not updated, just used to trigger sync sub-dir , so no need save to db

	RSYNC_LOG_PREFIX    = 12
	RSYNC_LOG_FILE_NAME = "rsync.log"

	RSYNC_TIMEOUT_SEC = "10"
)

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
	belogs.Debug("RsyncQuiet(): Command: rsync", "-Lirzts", "--del", "--timeout="+RSYNC_TIMEOUT_SEC, "--no-motd", "-4", rsyncUrl, rsyncDestPath)
	cmd := exec.Command("rsync", "-Lrzts", "--del", "--timeout="+RSYNC_TIMEOUT_SEC, "--no-motd", "-4", rsyncUrl, rsyncDestPath)
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
