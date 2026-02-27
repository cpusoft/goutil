package talutil

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/osutil"
)

type TalInfo struct {
	SyncUrl string `json:"syncUrl"`
	PubKey  string `json:"pubKey"`
}

func GetAllTalFile(file string) ([]string, error) {
	belogs.Notice("GetAllTalFile():input read file or path :", file)

	isDir, err := osutil.IsDir(file)
	if err != nil {
		belogs.Error("GetAllTalFile():IsDir err:", file, err)
		return nil, err
	}
	var files []string
	if isDir {
		suffixs := make(map[string]string)
		suffixs[".tal"] = ".tal"
		files, err = osutil.GetAllFilesBySuffixs(file, suffixs)
		if err != nil {
			belogs.Error("GetAllTalFile():GetAllFilesBySuffixs err:", file, err)
			return nil, err
		}
	} else {
		files = make([]string, 0)
		files = append(files, file)
	}
	belogs.Notice("GetAllTalFile(): files count: ", len(files))
	return files, nil
}

func ParseTalInfos(files []string) ([]TalInfo, error) {
	belogs.Debug("ParseTalInfos(): files:", len(files))

	talInfos := make([]TalInfo, 0)
	for _, v := range files {
		talInfo, err := parseTalInfo(v)
		if err != nil {
			belogs.Error("parseTalInfo(): file err: ", v, err)
			return nil, err
		}
		talInfos = append(talInfos, talInfo)
	}
	return talInfos, nil
}

func parseTalInfo(file string) (TalInfo, error) {
	belogs.Info("ParseTalInfo(): file:", file)

	talInfo := TalInfo{}
	f, err := os.Open(file)
	if err != nil {
		belogs.Error("ParseTalInfo(): file Open err:", file, err)
		return talInfo, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			belogs.Error("ParseTalInfo(): file Close err:", file, err)
		}
	}()

	input := bufio.NewScanner(f)
	// 修复：配置Scanner缓冲区，支持最大2MB单行（可根据需求调整）
	buf := make([]byte, 1024*1024) // 初始1MB缓冲区
	input.Buffer(buf, 2*1024*1024) // 最大2MB单行

	i := 0
	var buffer bytes.Buffer
	for input.Scan() {
		tmp := strings.TrimSpace(input.Text())
		if len(tmp) == 0 {
			continue
		}
		if i == 0 {
			tmp = strings.Replace(tmp, "\r\n", "", -1)
			talInfo.SyncUrl = tmp
		} else {
			buffer.WriteString(tmp)
		}
		i++
	}

	if err := input.Err(); err != nil {
		belogs.Error("ParseTalInfo(): scan file err:", file, err)
		return talInfo, err
	}

	talInfo.PubKey = buffer.String()
	belogs.Info("ParseTalInfo(): talInfo:", talInfo)
	return talInfo, nil
}
