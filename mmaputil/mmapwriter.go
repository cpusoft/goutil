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

func NewWriter() (mmap.MMap, *MmapHeader, error) {
	belogs.Debug("NewWriter(): gMmapConfig:", jsonutil.MarshalJson(gMmapConfig))
	if gMmapConfig == nil {
		belogs.Error("NewWriter(): gMmapConfig is empty")
		return nil, nil, errors.New("mmap is not init")
	}
	f, err := os.OpenFile(gMmapConfig.ShmFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		belogs.Error("NewWriter(): OpenFile fail", err)
		return nil, nil, err
	}
	defer f.Close()
	_ = f.Truncate(int64(gMmapConfig.TotalSize))

	mm, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		belogs.Error("NewWriter(): Map fail", err)
		return nil, nil, err
	}
	h := getHeader(mm)
	atomic.StoreUint32(&h.activeBuf, 0)
	atomic.StoreUint32(&h.usedA, 0)
	atomic.StoreUint32(&h.usedB, 0)
	atomic.StoreUint64(&h.version, 0)
	return mm, h, nil
}

// WriteStrings 核心函数：写入 []string，用 \n 分割每条
func WriteStrings(mm mmap.MMap, h *MmapHeader, items []string) {
	belogs.Debug("WriteStrings(): len(items)", len(items))
	if len(items) == 0 {
		belogs.Debug("WriteStrings(): items is empty")
		return
	}

	// 用换行符拼接所有字符串
	content := strings.Join(items, "\n") + "\n"
	data := []byte(content)
	dataLen := uint32(len(data))

	// 选择写入非活跃缓冲区
	active := atomic.LoadUint32(&h.activeBuf)
	var bufStart uint32
	var bufUsed *uint32

	if active == 0 {
		bufStart = uint32(gMmapConfig.BufBStart)
		bufUsed = &h.usedB
	} else {
		bufStart = uint32(gMmapConfig.BufAStart)
		bufUsed = &h.usedA
	}

	// 写入数据
	copy(mm[bufStart:bufStart+dataLen], data)

	// 更新长度、版本、切换缓冲区
	atomic.StoreUint32(bufUsed, dataLen)
	atomic.AddUint64(&h.version, 1)
	atomic.StoreUint32(&h.activeBuf, 1-active)

	belogs.Debug("WriteStrings(): ok, len(items):", len(items), "version:", atomic.LoadUint64(&h.version))
}
