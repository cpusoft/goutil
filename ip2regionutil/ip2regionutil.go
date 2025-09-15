package ip2regionutil

import (
	"errors"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type Ip2RegionModel struct {
	Ip       string `json:"ip"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	Isp      string `json:"isp"`
}

/*
Now: only search Ipv4 address from ip2region_v4.xdb file
*/
func SearchIp2Region(ipv4DataFilePathName, ip string) (Ip2RegionModel, error) {
	//var dbPath = `./ip2region.xdb`
	start := time.Now()
	belogs.Debug("SearchIp2Region(): ipv4DataFilePathName:", ipv4DataFilePathName, "  ip:", ip)
	version := xdb.IPv4
	searcher, err := xdb.NewWithFileOnly(version, ipv4DataFilePathName)
	if err != nil {
		belogs.Error("SearchIp2Region(): NewWithFileOnly fail, ipv4DataFilePathName:", ipv4DataFilePathName, err, "  time(s):", time.Since(start))
		return Ip2RegionModel{}, err
	}
	defer searcher.Close()

	// do the search
	region, err := searcher.SearchByStr(ip)
	if err != nil {
		belogs.Error("SearchIp2Region(): SearchByStr fail, ip:", ip, err, "  time(s):", time.Since(start))
		return Ip2RegionModel{}, err
	}

	// 中国|0|江苏省|南京市|0
	belogs.Debug("SearchIp2Region(): get region, ip:", ip, "   region:", region, time.Since(start))
	split := strings.Split(region, "|")
	if len(split) != 5 {
		belogs.Error("SearchIp2Region(): Split region fail, region:", region, "  time(s):", time.Since(start))
		return Ip2RegionModel{}, errors.New("region format error")
	}
	ip2RegionModel := Ip2RegionModel{
		Ip:       ip,
		Country:  removeZero(split[0]),
		Province: removeZero(split[2]),
		City:     removeZero(split[3]),
		Isp:      removeZero(split[4]),
	}

	belogs.Debug("SearchIp2Region(): get ip2RegionModel, ip", ip, "   ip2RegionModel:", jsonutil.MarshalJson(ip2RegionModel),
		time.Since(start))
	return ip2RegionModel, nil
}

func removeZero(s string) string {
	if s == "0" {
		return ""
	}
	return s
}
