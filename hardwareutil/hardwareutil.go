package hardwareutil

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

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
func GetDiskPartitions() ([]disk.PartitionStat, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		belogs.Error("GetDiskPartitions(): Partitions fail: ", err)
		return nil, err
	} else if len(parts) == 0 {
		belogs.Error("GetDiskPartitions():len(ret) == 0: ")
		return nil, errors.New("partitions is empty")
	}
	return parts, nil
}
func GetCpusInfo() ([]cpu.InfoStat, error) {
	cpus, err := cpu.Info()
	if err != nil {
		belogs.Error("GetCpusInfo(): Info fail: ", err)
		return nil, err
	} else if len(cpus) == 0 {
		belogs.Error("GetDiskPartitions():len(cpus) == 0: ")
		return nil, errors.New("cpu is empty")
	}
	return cpus, nil
}
