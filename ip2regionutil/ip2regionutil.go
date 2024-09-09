package ip2regionutil

import (
	"errors"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type IpRegionModel struct {
	Ip       string `json:"ip"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	Isp      string `json:"isp"`
}

func SearchIp2Region(dataFilePathName, ip string) (IpRegionModel, error) {
	//var dbPath = `./ip2region.xdb`
	start := time.Now()
	belogs.Debug("SearchIp2Region(): dataFilePathName:", dataFilePathName, "  ip:", ip)
	searcher, err := xdb.NewWithFileOnly(dataFilePathName)
	if err != nil {
		belogs.Error("SearchIp2Region(): NewWithFileOnly fail, dataFilePathName:", dataFilePathName, err, "  time(s):", time.Since(start))
		return IpRegionModel{}, err
	}
	defer searcher.Close()

	// do the search
	region, err := searcher.SearchByStr(ip)
	if err != nil {
		belogs.Error("SearchIp2Region(): SearchByStr fail, ip:", ip, err, "  time(s):", time.Since(start))
		return IpRegionModel{}, err
	}

	// 中国|0|江苏省|南京市|0
	belogs.Debug("SearchIp2Region(): get region, ip:", ip, "   region:", region, time.Since(start))
	split := strings.Split(region, "|")
	if len(split) != 5 {
		belogs.Error("SearchIp2Region(): Split region fail, region:", region, "  time(s):", time.Since(start))
		return IpRegionModel{}, errors.New("region format error")
	}
	ipRegionModel := IpRegionModel{
		Ip:       ip,
		Country:  split[0],
		Province: split[2],
		City:     split[3],
	}
	if split[4] != "0" {
		ipRegionModel.Isp = split[4]
	}
	belogs.Debug("SearchIp2Region(): get ipRegionModel, ip", ip, "   ipRegionModel:", jsonutil.MarshalJson(ipRegionModel),
		time.Since(start))
	return ipRegionModel, nil
}
