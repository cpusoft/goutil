package hardwareutil

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
)

func TestGetMemoryInfo(t *testing.T) {
	mem, _ := GetMemoryInfo()
	fmt.Println(mem)

	kernel, _ := GetKernelVersion()
	fmt.Println(kernel)

	host, _ := GetHostInfo()
	fmt.Println(host)

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

	cpus, _ := GetCpusInfo()
	for i := range cpus {
		fmt.Println(cpus[i])
	}
}
