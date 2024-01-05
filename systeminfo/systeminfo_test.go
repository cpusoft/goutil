package systeminfo

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
)

func TestGetMemoryInfo(t *testing.T) {
	mem, _ := GetMemoryInfo()
	fmt.Println("mem:\n", mem)

	kernel, _ := GetKernelVersion()
	fmt.Println("kernel:\n", kernel)

	host, _ := GetHostInfo()
	fmt.Println("host:\n", host)

	mis, avg, _ := GetProcesLoad()
	fmt.Println("host:mis\n", mis)
	fmt.Println("host:avg\n", avg)

	pars, _ := GetDiskPartitions()
	for i := range pars {
		fmt.Println(pars[i])
		serialNumber, _ := disk.SerialNumber(pars[i].Device)
		fmt.Println(serialNumber)
		/*  // only linux
		var st syscall.Stat_t
		syscall.Stat(pars[i].Device, &st)
		fmt.Printf("%+v\n", st)
		fmt.Println(jsonutil.MarshalJson(st))
		*/
	}
	usage, _ := GetDiskUsage()
	fmt.Println("usage\n", usage)

	cpus, _ := GetCpusInfo()
	fmt.Println("cpus")
	for i := range cpus {
		fmt.Println(cpus[i])
	}

	io, _ := GetNetIoCounter()
	fmt.Println("io\n", io)

	ifs, _ := GetNetInterfaces()
	fmt.Println("ifs\n", ifs)

	sys, err := GetSystemInfoUniqueIdSha256()
	fmt.Println(sys, err)
}
