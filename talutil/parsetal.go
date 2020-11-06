package talutil

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	osutil "github.com/cpusoft/goutil/osutil"
)

type TalInfo struct {
	SyncUrl string `json:"syncUrl"`
	PubKey  string `jons:"pubKey"`
}

func GetAllTalFile(file string) ([]string, error) {

	belogs.Notice("GetAllTalFile():input read file or path :", file)

	// 读取所有文件，加入到fileList列表中
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
	belogs.Info("ParseTalInfos(): files:", len(files))

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

	input := bufio.NewScanner(f)
	i := 0
	var buffer bytes.Buffer
	for input.Scan() { // 遇到 \n 或者\r\n循环一次
		tmp := strings.TrimSpace(input.Text())
		if i == 0 {
			tmp = strings.Replace(tmp, "\r\n", "", -1)
			talInfo.SyncUrl = tmp
		} else {
			buffer.WriteString(tmp)
		}
		i++
	}
	talInfo.PubKey = buffer.String()
	belogs.Info("ParseTalInfo(): talInfo:", talInfo)
	return talInfo, nil
}
