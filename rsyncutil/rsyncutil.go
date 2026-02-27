package rsyncutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/stringutil"
	"github.com/cpusoft/goutil/transportutil"
	"github.com/cpusoft/goutil/urlutil"
)

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
	belogs.Debug("RsyncQuiet():rsyncUrl:", rsyncUrl, " destPath:", destPath,
		"   will use  globalRsyncClientConfig:", jsonutil.MarshalJson(globalRsyncClientConfig))
	return RsyncQuietWithConfig(rsyncUrl, destPath, globalRsyncClientConfig)
}

// set rsync url and local dest path , then will call rsync
// will rsync every file but no output
// if success, the len(output) will be zero
func RsyncQuietWithConfig(rsyncUrl string, destPath string, rsyncClientConfig *RsyncClientConfig) (rsyncDestPath string, output []byte, err error) {
	belogs.Debug("RsyncQuietWithConfig():rsyncUrl:", rsyncUrl, " destPath:", destPath, "  rsyncClientConfig:", jsonutil.MarshalJson(rsyncClientConfig))
	defer func(rsyncUrl string) {
		if r := recover(); r != nil {
			errStack := string(debug.Stack())
			belogs.Error("RsyncQuietWithConfig(): recover from panic, rsyncUrl is :", rsyncUrl,
				" debug.Stack():", errStack, "  panic is :", r)
			// 修复1：将panic转为错误返回给调用方
			err = errors.New("panic in RsyncQuietWithConfig: " + fmt.Sprintf("%v", r) + ", stack: " + errStack)
		}
	}(rsyncUrl)

	// get host+path by url
	hostAndPath, err := urlutil.HostAndPath(rsyncUrl)
	belogs.Debug("RsyncQuietWithConfig():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath)
	if err != nil {
		belogs.Error("RsyncQuietWithConfig():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath, " err:", err)
		return "", output, err
	}

	// mkdirAll path
	rsyncDestPath = osutil.JoinPathFile(destPath, hostAndPath)
	belogs.Debug("RsyncQuietWithConfig():rsyncDestPath:", rsyncDestPath)
	err = os.MkdirAll(rsyncDestPath, os.ModePerm)
	if err != nil {
		belogs.Error("RsyncQuietWithConfig():MkdirAll:", rsyncDestPath, " err:", err)
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
	belogs.Debug("RsyncQuietWithConfig(): Command: rsync", "-Lirzts", "--del", "--timeout="+rsyncClientConfig.Timeout, "--contimeout="+rsyncClientConfig.ConTimeout,
		"--no-motd", "-4", rsyncUrl, rsyncDestPath)
	cmd := exec.Command("rsync", "-Lrzts", "--del", "--timeout="+rsyncClientConfig.Timeout, "--contimeout="+rsyncClientConfig.ConTimeout,
		"--no-motd", "-4", rsyncUrl, rsyncDestPath)
	// if success, the len(output) will be zero
	output, err = cmd.CombinedOutput()
	if err != nil {
		belogs.Error("RsyncQuietWithConfig(): exec.Command fail, rsyncUrl is :", rsyncUrl, "   output is ", string(output), " err is :", err)
		// some err detail in output
		err = errors.New(string(output) + ", " + err.Error())
		return "", output, err
	}
	belogs.Debug("RsyncQuietWithConfig(): rsyncDestPath:", rsyncDestPath, "  output:", string(output))
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
	// 修复2：移除--log-file参数的手动引号（exec.Command会自动处理参数分隔）
	belogs.Debug("RsyncToLogFile(): Command: rsync", "-Lirzts", "--del", "--no-motd", "-4", "--log-file="+rsyncLogFile, rsyncUrl, rsyncDestPath)
	cmd := exec.Command("rsync", "-Lirzts", "--del", "--no-motd", "-4", "--log-file="+rsyncLogFile, rsyncUrl, rsyncDestPath)
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

	// 移除多余的初始化：output = make([]byte, 0)

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
	// 修复3：使用全局配置的timeout，而非硬编码6000
	timeout := globalRsyncClientConfig.Timeout
	belogs.Debug("RsyncToStdout(): Command: rsync", "-Lirzts", "--del", "--timeout="+timeout, "--no-motd", "-4", rsyncUrl, rsyncDestPath)
	cmd := exec.Command("rsync", "-Lirzts", "--del", "--timeout="+timeout, "--no-motd", "-4", rsyncUrl, rsyncDestPath)
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

/*
https://stackoverflow.com/questions/40612505/format-of-rsync-logfile
  - : delete
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
		result = stringutil.TrimNewLine(result)
		// 关键修复：跳过空行（分割stdout时可能产生）
		if result == "" {
			continue
		}
		first := result[0]
		second := result[1]
		rsyncResult := RsyncResult{}

		switch first {
		case '*':
			rsyncResult.RsyncType = RSYNC_TYPE_DEL
			// 关键修复：截取从第12位开始的完整文件名，而非硬编码RSYNC_LOG_PREFIX（避免截取不全）
			fileNamePart := strings.TrimSpace(result[RSYNC_LOG_PREFIX:])
			if fileNamePart == "" {
				continue // 空文件名跳过
			}
			rsyncResult.FilePath, rsyncResult.FileName = osutil.GetFilePathAndFileName(rsyncDestPath + fileNamePart)
			rsyncResult.FileType = strings.Replace(path.Ext(rsyncResult.FileName), ".", "", -1)
		case '>':
			if strings.Contains(result, "++++") {
				rsyncResult.RsyncType = RSYNC_TYPE_ADD
			} else {
				rsyncResult.RsyncType = RSYNC_TYPE_UPDATE
			}
			fileNamePart := strings.TrimSpace(result[RSYNC_LOG_PREFIX:])
			if fileNamePart == "" {
				continue
			}
			rsyncResult.FilePath, rsyncResult.FileName = osutil.GetFilePathAndFileName(rsyncDestPath + fileNamePart)
			rsyncResult.FileType = strings.Replace(path.Ext(rsyncResult.FileName), ".", "", -1)
		case 'c':
			switch second {
			case 'd':
				rsyncResult.RsyncType = RSYNC_TYPE_MKDIR
				dirNamePart := strings.TrimSpace(result[RSYNC_LOG_PREFIX:])
				if dirNamePart == "" {
					continue
				}
				rsyncResult.FilePath = rsyncDestPath + dirNamePart
				// MKDIR类型FileName为空，需手动设置（避免被过滤）
				rsyncResult.FileName = filepath.Base(dirNamePart)
			default:
				rsyncResult.RsyncType = RSYNC_TYPE_IGNORE
			}
		case '<', 'L', 'h', '.':
			fallthrough
		default:
			rsyncResult.RsyncType = RSYNC_TYPE_IGNORE
		}

		rsyncResult.SyncTime = time.Now()
		fullPath := rsyncResult.FilePath + rsyncResult.FileName
		isDir, dirErr := osutil.IsDir(fullPath)
		if dirErr != nil {
			belogs.Error("parseToRsyncResults(): osutil.IsDir fail, fullPath:", fullPath, " err:", dirErr)
		}
		rsyncResult.IsDir = isDir

		belogs.Info("parseToRsyncResults():rsyncResult:", jsonutil.MarshalJson(rsyncResult))
		// 关键修复：MKDIR类型允许FileName为空（因为是目录）
		if rsyncResult.RsyncType == RSYNC_TYPE_IGNORE ||
			(rsyncResult.RsyncType != RSYNC_TYPE_MKDIR && (len(rsyncResult.FileName) == 0 || len(rsyncResult.FilePath) == 0)) {
			belogs.Debug("parseToRsyncResults():ignore parseRsyncResult: ", rsyncResult)
			continue
		}
		rsyncResult.RsyncUrl = rsyncUrl
		rsyncResults = append(rsyncResults, rsyncResult)
	}
	belogs.Debug("parseToRsyncResults():rsyncResults: ", jsonutil.MarshalJson(rsyncResults))
	return rsyncResults, nil
}

// need read all current existed cer file, to just to trigger sub ca repo sync
// 同时修复AddCerToRsyncResults中文件名对比逻辑（确保GetFilesInDir返回纯文件名）
func AddCerToRsyncResults(rsyncDestPath string, rsyncResults []RsyncResult) (err error) {
	// 修复：existingFiles 存储 完整路径+文件名，避免纯文件名匹配错误
	existingFiles := make(map[string]bool, len(rsyncResults))
	for _, rsyncResult := range rsyncResults {
		fullName := osutil.JoinPathFile(rsyncResult.FilePath, rsyncResult.FileName)
		existingFiles[fullName] = true
	}

	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	files, err := osutil.GetFilesInDir(rsyncDestPath, m)
	if err != nil {
		belogs.Debug("addCerToRsyncResults():GetFilesInDir, files:", files, err)
		return err
	}
	belogs.Debug("addCerToRsyncResults():GetFilesInDir, files:", files)
	for _, file := range files {
		fullFile := osutil.JoinPathFile(rsyncDestPath, file)
		belogs.Debug("addCerToRsyncResults(): check file:", fullFile)
		if !existingFiles[fullFile] { // 对比完整路径
			// 原有逻辑不变
			rsyncResult := RsyncResult{}
			rsyncResult.RsyncType = RSYNC_TYPE_JUST_SYNC
			rsyncResult.FilePath = rsyncDestPath
			rsyncResult.FileName = file
			rsyncResult.FileType = "cer"
			rsyncResult.SyncTime = time.Now()
			fullFilePath := osutil.JoinPathFile(rsyncDestPath, file)
			isDir, dirErr := osutil.IsDir(fullFilePath)
			if dirErr != nil {
				belogs.Error("addCerToRsyncResults(): osutil.IsDir fail, fullFilePath:", fullFilePath, " err:", dirErr)
				continue
			}
			rsyncResult.IsDir = isDir
			rsyncResults = append(rsyncResults, rsyncResult)
			belogs.Info("addCerToRsyncResults(): manual add rsyncResult:", jsonutil.MarshalJson(rsyncResult))
		}
	}
	belogs.Debug("addCerToRsyncResults(): after add cer, rsyncResults:", jsonutil.MarshalJson(rsyncResults))
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
	// 修复7：创建filesFromDisk副本，避免修改入参map
	filesFromDiskCopy := make(map[string]RsyncFileHash, len(filesFromDisk))
	for k, v := range filesFromDisk {
		filesFromDiskCopy[k] = v
	}

	// if db is empty, so all filesFromDisk is add
	if len(filesFromDb) == 0 {
		return filesFromDisk, nil, nil, nil, nil
	}

	// if disk is empty, so all filesFromDb is del
	if len(filesFromDiskCopy) == 0 {
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
		if valueDisk, ok := filesFromDiskCopy[keyDb]; ok {
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
			delete(filesFromDiskCopy, keyDb)
		} else {
			// if not found in disk ,then is del, so save to delFiles, and value is db
			delFiles[keyDb] = valueDb
		}
	}
	addFiles = filesFromDiskCopy
	belogs.Debug("DiffFiles(): len(addFiles):", len(addFiles), jsonutil.MarshalJson(addFiles))
	belogs.Debug("DiffFiles(): len(delFiles):", len(delFiles), jsonutil.MarshalJson(delFiles))
	belogs.Debug("DiffFiles(): len(updateFiles):", len(updateFiles), jsonutil.MarshalJson(updateFiles))
	belogs.Debug("DiffFiles(): len(noChangeFiles):", len(noChangeFiles), jsonutil.MarshalJson(noChangeFiles))
	belogs.Info("DiffFiles(): len(addFiles):", len(addFiles), "  len(delFiles):", len(delFiles),
		"  len(updateFiles):", len(updateFiles), "  len(noChangeFiles):", len(noChangeFiles), "  time(s):", time.Since(start))

	return addFiles, delFiles, updateFiles, noChangeFiles, nil
}
