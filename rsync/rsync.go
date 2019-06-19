package rsync

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	belogs "github.com/astaxie/beego/logs"
	osutil "github.com/cpusoft/goutil/osutil"
	urlutil "github.com/cpusoft/goutil/urlutil"
)

// rsync type
const (
	RSYNC_TYPE_ADD    = "add"
	RSYNC_TYPE_DEL    = "del"
	RSYNC_TYPE_UPDATE = "update"
	RSYNC_TYPE_MKDIR  = "mkdir"
	RSYNC_TYPE_IGNORE = "ignore"

	RSYNC_LOG_PREFIX = 12
)

type RsyncRecord struct {
	Id           uint64        `json:"id"`
	StartTime    time.Time     `json:"startTime"`
	EndTime      time.Time     `json:"endTime"`
	Style        string        `json:"style"`
	RsyncResults []RsyncResult `json:"rsyncResults"`
}

type RsyncResult struct {
	Id        uint64    `json:"id"`
	RsyncId   uint64    `json:"rsyncId"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	FileType  string    `json:"fileType"`
	RsyncType string    `json:"rsyncType"`
	IsDir     bool      `json:"isDir"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// set rsync url and local dest path , then will call rsync
// will get all stdout to get every file rsync state
func Rsync(rsyncUrl string, destPath string) ([]RsyncResult, error) {
	belogs.Debug("Rsync():rsyncUrl:", rsyncUrl, " destPath:", destPath)

	rysncResults := make([]RsyncResult, 0)

	// mkdirAll path
	err := os.MkdirAll(destPath, os.ModePerm)
	if err != nil {
		belogs.Error("Rsync():MkdirAll:", destPath, " err:", err)
		return rysncResults, err
	}

	// get host+path by url
	hostAndPath, err := urlutil.HostAndPath(rsyncUrl)
	belogs.Debug("Rsync():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath)
	if err != nil {
		belogs.Error("Rsync():HostAndPath: rsyncUrl:", rsyncUrl, "  HostAndPath:", hostAndPath, " err:", err)
		return rysncResults, err
	}

	// call rsync
	//rsync -Lirzts --del --timeout=5 --contimeout=5 --no-motd  -4 rsync://rpki.afrinic.net/repository/afrinic/  /tmp/rpki.afrinic.net/repository/afrinic/
	cmd := exec.Command("rsync", "-Lirzts", "--del", "--timeout=15", "--contimeout=15", "--no-motd", "-4", rsyncUrl, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belogs.Error("Rsync(): exec.Command: err: ", err, ": "+string(output))
		return nil, err
	}

	// if no changed ,so length of output is 0
	if len(output) == 0 {
		return rysncResults, nil
	}

	result := string(output)
	results := strings.Split(result, "\r\n")
	belogs.Debug("Rsync(): len(results):", len(results))
	for _, one := range results {
		if len(one) <= RSYNC_LOG_PREFIX {
			continue
		}
		one = strings.Replace(one, "\n", "", -1)
		one = strings.Replace(one, "\r", "", -1)
		rsyncResult, err := parseRsyncResult(destPath+hostAndPath, one)
		if err != nil {
			belogs.Error("Rsync(): parseRsyncResult: err: ", err, ": "+one)
			return rysncResults, err
		}
		rysncResults = append(rysncResults, rsyncResult)
	}

	return rysncResults, nil
}

func parseRsyncResult(destPath, result string) (RsyncResult, error) {
	belogs.Debug("parseRsyncResult():destPath:", destPath, " result:", result)

	if !strings.HasSuffix(destPath, string(os.PathSeparator)) {
		destPath = destPath + string(os.PathSeparator)
	}
	first := result[0]
	second := result[1]
	rsyncResult := RsyncResult{}
	/*

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
			osutil.GetFilePathAndFileName(destPath + string(result[RSYNC_LOG_PREFIX:]))
		rsyncResult.FileType = path.Ext(rsyncResult.FileName)
	case '>':
		if strings.Contains(result, "++++") {
			rsyncResult.RsyncType = RSYNC_TYPE_ADD
		} else {
			rsyncResult.RsyncType = RSYNC_TYPE_UPDATE
		}
		rsyncResult.FilePath, rsyncResult.FileName =
			osutil.GetFilePathAndFileName(destPath + string(result[RSYNC_LOG_PREFIX:]))
		rsyncResult.FileType = path.Ext(rsyncResult.FileName)
	case 'c':
		switch second {
		case 'd':
			rsyncResult.RsyncType = RSYNC_TYPE_MKDIR
			rsyncResult.FilePath = destPath + string(result[RSYNC_LOG_PREFIX:])

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
	belogs.Debug("parseRsyncResult():rsyncResult:", rsyncResult)
	return rsyncResult, nil

}
