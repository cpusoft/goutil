// systeminfo_test.go (无Mock，纯真实环境测试)
package systeminfo

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/stretchr/testify/assert"
)

// -------------------------- 正常场景测试（核心函数） --------------------------
// 测试GetMemoryInfo：验证返回非空且无错误
func TestGetMemoryInfo_Normal(t *testing.T) {
	stat, err := GetMemoryInfo()
	assert.NoError(t, err, "获取内存信息应无错误")
	assert.NotNil(t, stat, "内存信息返回值不应为nil")
	assert.Greater(t, stat.Total, uint64(0), "内存总容量应大于0")
}

// 测试GetKernelVersion：验证返回非空版本号
func TestGetKernelVersion_Normal(t *testing.T) {
	version, err := GetKernelVersion()
	assert.NoError(t, err, "获取内核版本应无错误")
	assert.NotEmpty(t, version, "内核版本号不应为空")
}

// 测试GetHostInfo：验证主机信息关键字段非空
func TestGetHostInfo_Normal(t *testing.T) {
	info, err := GetHostInfo()
	assert.NoError(t, err, "获取主机信息应无错误")
	assert.NotNil(t, info, "主机信息返回值不应为nil")
	assert.NotEmpty(t, info.OS, "操作系统名称不应为空")
	assert.NotEmpty(t, info.KernelArch, "内核架构不应为空")
	assert.NotEmpty(t, info.HostID, "主机ID不应为空")
}

// 测试GetProcesLoad：验证返回值（允许部分字段为空，但无错误）
func TestGetProcesLoad_Normal(t *testing.T) {
	miscStat, avgStat, err := GetProcesLoad()
	assert.NoError(t, err, "获取负载信息应无错误")
	// 不同系统的LoadAvg可能返回nil（如Windows），做兼容判断
	if runtime.GOOS != "windows" {
		assert.NotNil(t, avgStat, "非Windows系统LoadAvg不应为nil")
	}
	// MiscStat在所有系统应返回非nil
	assert.NotNil(t, miscStat, "MiscStat返回值不应为nil")
}

// 测试GetDiskPartitions：验证返回非空分区列表
func TestGetDiskPartitions_Normal(t *testing.T) {
	parts, err := GetDiskPartitions()
	assert.NoError(t, err, "获取磁盘分区应无错误")
	assert.NotEmpty(t, parts, "磁盘分区列表不应为空")
	// 验证分区的挂载点和文件系统字段
	for _, part := range parts {
		assert.NotEmpty(t, part.Mountpoint, "分区挂载点不应为空")
		assert.NotEmpty(t, part.Fstype, "分区文件系统类型不应为空")
	}
}

// 测试GetDiskUsage：验证指定路径的磁盘使用信息
func TestGetDiskUsage_Normal(t *testing.T) {
	usage, err := GetDiskUsage()
	assert.NoError(t, err, "获取磁盘使用信息应无错误")
	assert.NotNil(t, usage, "磁盘使用信息返回值不应为nil")
	assert.Greater(t, usage.Total, uint64(0), "磁盘总容量应大于0")
}

// 测试GetCpusInfo：验证CPU信息非空且关键字段有效
func TestGetCpusInfo_Normal(t *testing.T) {
	cpus, err := GetCpusInfo()
	assert.NoError(t, err, "获取CPU信息应无错误")
	assert.NotEmpty(t, cpus, "CPU信息列表不应为空")
	// 验证第一个CPU的关键字段
	firstCPU := cpus[0]
	assert.NotEmpty(t, firstCPU.VendorID, "CPU厂商ID不应为空")
	assert.NotEmpty(t, firstCPU.ModelName, "CPU型号名称不应为空")
	assert.Greater(t, firstCPU.Cores, int32(0), "CPU核心数应大于0")
}

// 测试GetNetIoCounter：验证网络IO统计（兼容多系统）
func TestGetNetIoCounter_Normal(t *testing.T) {
	io, err := GetNetIoCounter()
	// 部分系统（如容器环境）可能返回非1长度的IO列表，此处兼容判断
	if err != nil {
		assert.Equal(t, "ioCounters is not summary", err.Error(), "错误信息应匹配")
		t.Log("当前系统NetIO计数器长度非1，跳过返回值验证")
		return
	}
	assert.NotNil(t, io, "网络IO统计返回值不应为nil")
	assert.GreaterOrEqual(t, io.BytesRecv, uint64(0), "接收字节数应≥0")
	assert.GreaterOrEqual(t, io.BytesSent, uint64(0), "发送字节数应≥0")
}

// 测试GetNetInterfaces：验证网络接口列表非空
func TestGetNetInterfaces_Normal(t *testing.T) {
	ifs, err := GetNetInterfaces()
	assert.NoError(t, err, "获取网络接口应无错误")
	assert.NotEmpty(t, ifs, "网络接口列表不应为空")
	// 验证至少有一个接口有名称
	hasValidInterface := false
	for _, iface := range ifs {
		if iface.Name != "" {
			hasValidInterface = true
			break
		}
	}
	assert.True(t, hasValidInterface, "至少有一个网络接口名称非空")
}

// 测试GetSystemInfoUniqueId：验证唯一ID所有字段非空
func TestGetSystemInfoUniqueId_Normal(t *testing.T) {
	uniqueId, err := GetSystemInfoUniqueId()
	assert.NoError(t, err, "生成系统唯一ID应无错误")
	// 验证所有字段非空
	assert.NotEmpty(t, uniqueId.HostOs, "HostOs不应为空")
	assert.NotEmpty(t, uniqueId.HostKernelArch, "HostKernelArch不应为空")
	assert.NotEmpty(t, uniqueId.HostId, "HostId不应为空")
	assert.NotEmpty(t, uniqueId.CpuVendorId, "CpuVendorId不应为空")
	assert.NotEmpty(t, uniqueId.CpuPhysicalId, "CpuPhysicalId不应为空")
	assert.NotEmpty(t, uniqueId.CpuModelName, "CpuModelName不应为空")
	assert.NotEmpty(t, uniqueId.MemoryTotal, "MemoryTotal不应为空")
}

// 测试GetSystemInfoUniqueIdSha256：验证SHA256哈希长度为32字节
func TestGetSystemInfoUniqueIdSha256_Normal(t *testing.T) {
	sha256Bytes, err := GetSystemInfoUniqueIdSha256()
	assert.NoError(t, err, "生成SHA256哈希应无错误")
	assert.Len(t, sha256Bytes, 32, "SHA256哈希长度必须为32字节")
	assert.NotZero(t, sha256Bytes, "SHA256哈希不应为全0")
}

// -------------------------- 临界场景测试（边界条件） --------------------------
// 测试GetDiskPartitions：模拟空分区（注：真实系统几乎不可能，验证错误逻辑）
// 说明：真实系统无法构造空分区，此处通过直接调用disk.Partitions模拟临界逻辑
func TestGetDiskPartitions_Empty_Critical(t *testing.T) {
	// 调用底层函数，传入all=true（可能返回空列表）
	parts, _ := disk.Partitions(true)
	if len(parts) == 0 {
		// 若真为空，验证GetDiskPartitions的错误处理
		_, err := GetDiskPartitions()
		assert.Error(t, err)
		assert.Equal(t, "partitions is empty", err.Error())
	} else {
		t.Log("当前系统磁盘分区非空，跳过空分区临界测试")
	}
}

// 测试GetCpusInfo：模拟空CPU列表（验证错误逻辑）
func TestGetCpusInfo_Empty_Critical(t *testing.T) {
	// 调用底层函数，验证返回值
	cpus, _ := cpu.Info()
	if len(cpus) == 0 {
		_, err := GetCpusInfo()
		assert.Error(t, err)
		assert.Equal(t, "cpu is empty", err.Error())
	} else {
		t.Log("当前系统CPU信息非空，跳过空CPU临界测试")
	}
}

// 测试GetNetIoCounter：临界场景（长度≠1）
func TestGetNetIoCounter_InvalidLength_Critical(t *testing.T) {
	ioCounters, err := net.IOCounters(false)
	assert.NoError(t, err)
	// 若长度≠1，验证错误返回
	if len(ioCounters) != 1 {
		_, err := GetNetIoCounter()
		assert.Error(t, err)
		assert.Equal(t, "ioCounters is not summary", err.Error())
	} else {
		t.Log("当前系统NetIO计数器长度为1，跳过长度异常临界测试")
	}
}

// 测试GetSystemInfoUniqueId：依赖函数失败的临界场景
func TestGetSystemInfoUniqueId_DependFail_Critical(t *testing.T) {
	// 模拟主机信息获取失败（注：真实系统无法构造，验证错误传递逻辑）
	// 此处通过直接调用host.Info验证错误传递
	_, err := host.Info()
	if err != nil {
		_, err := GetSystemInfoUniqueId()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GetHostInfo fail")
	} else {
		t.Log("当前系统主机信息获取正常，跳过依赖失败临界测试")
	}
}

// -------------------------- 性能基准测试（核心函数） --------------------------
// 基准测试：GetMemoryInfo执行效率
func BenchmarkGetMemoryInfo(b *testing.B) {
	// 重置计时器，排除初始化耗时
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetMemoryInfo()
	}
}

// 基准测试：GetSystemInfoUniqueId（核心唯一ID生成）
func BenchmarkGetSystemInfoUniqueId(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetSystemInfoUniqueId()
	}
}

// 基准测试：GetSystemInfoUniqueIdSha256（哈希计算）
func BenchmarkGetSystemInfoUniqueIdSha256(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetSystemInfoUniqueIdSha256()
	}
}

// 基准测试：GetCpusInfo（CPU信息获取）
func BenchmarkGetCpusInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetCpusInfo()
	}
}

// 基准测试：GetDiskUsage（磁盘使用信息获取）
func BenchmarkGetDiskUsage(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetDiskUsage()
	}
}

// ////////////////////////////////////////////////////////////
// //////////////////////////////////////
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
