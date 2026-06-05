package mmaputil

import (
	"unsafe"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	mmap "github.com/edsrzf/mmap-go"
)

type MmapConfig struct {
	ShmFile    string `json:"shmFile"`    //  = "./shm_double_buf.dat"
	HeaderSize int    `json:"headerSize"` // = 256
	MaxBufSize int    `json:"maxBufSize"` // = 1024 * 1024 * 1024
	TotalSize  int    `json:"totalSize"`  // = HeaderSize + MaxBufSize*2
	BufAStart  int    `json:"bufAStart"`  // = HeaderSize
	BufBStart  int    `json:"bufBStart"`  // = HeaderSize + MaxBufSize
}

var gMmapConfig *MmapConfig

func init() {
	gMmapConfig = &MmapConfig{
		ShmFile:    conf.DefaultString("mmap::shmFile", "/tmp/shm.dat"),
		HeaderSize: conf.DefaultInt("mmap::headerSize", 256),
		MaxBufSize: conf.DefaultInt("mmap::itemSize", 1024*1024*1024),
	}
	gMmapConfig.TotalSize = gMmapConfig.HeaderSize + gMmapConfig.MaxBufSize*2
	gMmapConfig.BufAStart = gMmapConfig.HeaderSize
	gMmapConfig.BufBStart = gMmapConfig.HeaderSize + gMmapConfig.MaxBufSize
	belogs.Debug("mmap.init(): gMmapConfig", jsonutil.MarshalJson(gMmapConfig))
}

type MmapHeader struct {
	activeBuf uint32
	usedA     uint32
	usedB     uint32
	version   uint64
	_         [228]byte
}

func getHeader(mm mmap.MMap) *MmapHeader {
	return (*MmapHeader)(unsafe.Pointer(&mm[0]))
}
