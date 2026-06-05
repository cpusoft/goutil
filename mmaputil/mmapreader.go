package mmaputil

import (
	"errors"
	"os"
	"strings"
	"sync/atomic"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	mmap "github.com/edsrzf/mmap-go"
)

func NewReader() (mmap.MMap, *MmapHeader, error) {
	belogs.Debug("NewReader(): gMmapConfig:", jsonutil.MarshalJson(gMmapConfig))
	if gMmapConfig == nil {
		belogs.Error("NewReader(): gMmapConfig is empty")
		return nil, nil, errors.New("mmap is not init")
	}
	f, err := os.OpenFile(gMmapConfig.ShmFile, os.O_RDONLY, 0644)
	if err != nil {
		belogs.Error("NewReader(): OpenFile fail", err)
		return nil, nil, err
	}
	defer f.Close()

	mm, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		belogs.Error("NewReader(): Map fail", err)
		return nil, nil, err
	}
	return mm, getHeader(mm), nil
}

// ReadStrings 读取并按 \n 分割成字符串数组
func ReadStrings(mm mmap.MMap, h *MmapHeader) []string {
	active := atomic.LoadUint32(&h.activeBuf)
	var bufStart uint32
	var bufLen uint32

	if active == 0 {
		bufStart = uint32(gMmapConfig.BufAStart)
		bufLen = atomic.LoadUint32(&h.usedA)
	} else {
		bufStart = uint32(gMmapConfig.BufBStart)
		bufLen = atomic.LoadUint32(&h.usedB)
	}

	if bufLen == 0 {
		belogs.Error("ReadStrings(): bufLen is empty")
		return nil
	}

	// 读取完整块
	data := mm[bufStart : bufStart+bufLen]

	// 按换行符分割
	lines := strings.Split(string(data), "\n")

	// 过滤空行
	var res []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			res = append(res, line) // 保留原始行
		}
	}
	belogs.Debug("ReadStrings(): ok, len(res)", len(res))
	return res
}
