package systeminfo

import (
	"crypto/sha256"
	"errors"
	"runtime"

	"github.com/cpusoft/goutil/base64util"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type SystemInfoUniqueId struct {
	HostOs         string `json:"hostOs"`
	HostKernelArch string `json:"hostKernelArch"`
	HostId         string `json:"hostId"`
	CpuVendorId    string `json:"cpuVendorId"`
	CpuPhysicalId  string `json:"cpuPhysicalId"`
	CpuModelName   string `json:"cpuModelName"`
	MemoryTotal    string `json:"memoryTotal"`
}

func GetMemoryInfo() (*mem.VirtualMemoryStat, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		belogs.Error("GetMemoryInfo():VirtualMemory fail: ", err)
		return nil, err
	}
	return v, nil
}

func GetKernelVersion() (string, error) {
	version, err := host.KernelVersion()
	if err != nil {
		belogs.Error("GetKernelVersion(): KernelVersion fail: ", err)
		return "", err
	}
	return version, nil
}

func GetHostInfo() (*host.InfoStat, error) {
	hostInfo, err := host.Info()
	if err != nil {
		belogs.Error("GetHostInfo(): Info fail: ", err)
		return nil, err
	}
	return hostInfo, nil
}

func GetProcesLoad() (*load.MiscStat, *load.AvgStat, error) {
	misStat, err := load.Misc()
	if err != nil {
		belogs.Error("GetProcesLoad(): Misc fail: ", err)
	}

	avgStat, err := load.Avg()
	if err != nil {
		belogs.Error("GetProcesLoad(): Avg fail: ", err)
	}

	return misStat, avgStat, nil
}

func GetDiskPartitions() ([]disk.PartitionStat, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		belogs.Error("GetDiskPartitions(): Partitions fail: ", err)
		return nil, err
	} else if len(parts) == 0 {
		belogs.Error("GetDiskPartitions():len(parts) == 0: no disk partitions found")
		return nil, errors.New("partitions is empty")
	}
	return parts, nil
}

func GetDiskUsage() (*disk.UsageStat, error) {
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	u, err := disk.Usage(path)
	if err != nil {
		belogs.Error("GetDiskUsage(): Usage fail for path '", path, "': ", err)
		return nil, err
	}
	return u, nil
}

func GetCpusInfo() ([]cpu.InfoStat, error) {
	cpus, err := cpu.Info()
	if err != nil {
		belogs.Error("GetCpusInfo(): Info fail: ", err)
		return nil, err
	} else if len(cpus) == 0 {
		belogs.Error("GetCpusInfo():len(cpus) == 0: no CPU info found")
		return nil, errors.New("cpu is empty")
	}
	return cpus, nil
}

func GetNetIoCounter() (*net.IOCountersStat, error) {
	ioCounters, err := net.IOCounters(false)
	if err != nil {
		belogs.Error("GetNetIoCounter(): IOCounters fail: ", err)
		return nil, err
	}
	if len(ioCounters) != 1 {
		belogs.Error("GetNetIoCounter(): len(ioCounters) =", len(ioCounters), " (expected 1)")
		return nil, errors.New("ioCounters is not summary")
	}
	v := ioCounters[0]
	return &v, nil
}

func GetNetInterfaces() (net.InterfaceStatList, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		belogs.Error("GetNetInterfaces(): Interfaces fail: ", err)
		return nil, err
	}
	return ifs, nil
}

func GetSystemInfoUniqueId() (systemInfoUniqueId SystemInfoUniqueId, err error) {
	hostInfo, err := GetHostInfo()
	if err != nil {
		belogs.Error("GetSystemInfoUniqueId(): GetHostInfo fail: ", err)
		return systemInfoUniqueId, err
	}
	// 改用jsonutil.MarshalJson，不校验返回值（按你的要求）
	belogs.Debug("GetSystemInfoUniqueId(): GetHostInfo hostInfo: ", jsonutil.MarshalJson(hostInfo))

	cpusInfo, err := GetCpusInfo()
	if err != nil {
		belogs.Error("GetSystemInfoUniqueId(): GetCpusInfo fail: ", err)
		return systemInfoUniqueId, err
	}
	// 防御性检查：避免极端情况下切片索引越界
	if len(cpusInfo) == 0 {
		belogs.Error("GetSystemInfoUniqueId(): cpusInfo is empty unexpectedly")
		return systemInfoUniqueId, errors.New("cpu info is empty")
	}
	// 改用jsonutil.MarshalJson，不校验返回值
	belogs.Debug("GetSystemInfoUniqueId(): GetHostInfo cpusInfo: ", jsonutil.MarshalJson(cpusInfo))

	memoryInfo, err := GetMemoryInfo()
	if err != nil {
		belogs.Error("GetSystemInfoUniqueId(): GetMemoryInfo fail: ", err)
		return systemInfoUniqueId, err
	}
	// 改用jsonutil.MarshalJson，不校验返回值
	belogs.Debug("GetSystemInfoUniqueId(): GetMemoryInfo memoryInfo: ", jsonutil.MarshalJson(memoryInfo))

	systemInfoUniqueId = SystemInfoUniqueId{
		HostOs:         hostInfo.OS,
		HostKernelArch: hostInfo.KernelArch,
		HostId:         hostInfo.HostID,
		CpuVendorId:    cpusInfo[0].VendorID,
		CpuPhysicalId:  cpusInfo[0].PhysicalID,
		CpuModelName:   cpusInfo[0].ModelName,
		MemoryTotal:    convert.ToString(memoryInfo.Total),
	}
	// 改用jsonutil.MarshalJson，不校验返回值
	belogs.Info("GetSystemInfoUniqueId(): systemInfoUniqueId: ", jsonutil.MarshalJson(systemInfoUniqueId))
	return systemInfoUniqueId, nil
}

func GetSystemInfoUniqueIdSha256() (systemInfoUniqueIdSha256 []byte, err error) {
	systemInfoUniqueId, err := GetSystemInfoUniqueId()
	if err != nil {
		belogs.Error("GetSystemInfoUniqueIdSha256(): GetSystemInfoUniqueId fail: ", err)
		return nil, err
	}
	// 改用jsonutil.MarshalJson，不校验返回值（按你的要求）
	systemInfoUniqueIdData := []byte(jsonutil.MarshalJson(systemInfoUniqueId))
	systemInfoUniqueIdSha256Data := sha256.Sum256(systemInfoUniqueIdData)
	systemInfoUniqueIdSha256 = systemInfoUniqueIdSha256Data[:]

	// 保留敏感日志修复：移除明文数据，降级为Debug级别
	belogs.Debug("GetSystemInfoUniqueIdSha256(): systemInfoUniqueIdSha256 base64:", base64util.EncodeBase64(systemInfoUniqueIdSha256),
		"    systemInfoUniqueIdSha256 hex:", convert.PrintBytesOneLine(systemInfoUniqueIdSha256))
	return systemInfoUniqueIdSha256, nil
}
