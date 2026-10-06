package langutil

import (
	"unicode/utf8"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// 包全局单例detector，并发安全，避免每次调用新建对象
var detector = chardet.NewTextDetector()

// codecCandidate 编码候选项，统一结构体，消除重复定义
type codecCandidate struct {
	name string
	dec  encoding.Encoding
}

// TryDecodeToUTF8 自动探测编码并转为UTF8
// 优先检查原始字节是否合法UTF8；否则chardet探测，探测失败则使用预设候选编码依次尝试
// 最终兜底清洗非法UTF8字节，保证输出一定合法UTF8，用于RIR Whois采集
// 返回值：result=utf8字符串，usedCodec=本次使用的编码标识(utf8/cp1252/latin1/fallback-clean等，用于日志统计)
func TryDecodeToUTF8(raw []byte) (result string, usedCodec string) {
	if utf8.Valid(raw) {
		return string(raw), "utf8"
	}

	detRes, err := detector.DetectBest(raw)
	var candidates []codecCandidate

	// chardet探测成功，映射对应解码器
	if err == nil && detRes != nil && len(detRes.Charset) > 0 {
		switch detRes.Charset {
		case "Windows-1252":
			candidates = append(candidates, codecCandidate{"cp1252", charmap.Windows1252})
		case "ISO-8859-1", "LATIN1":
			candidates = append(candidates, codecCandidate{"latin1", charmap.ISO8859_1})
		case "ISO-8859-15":
			candidates = append(candidates, codecCandidate{"iso8859-15", charmap.ISO8859_15})
		case "ISO-8859-2":
			candidates = append(candidates, codecCandidate{"iso8859-2", charmap.ISO8859_2})
		case "KOI8-R":
			candidates = append(candidates, codecCandidate{"koi8r", charmap.KOI8R})
		case "GBK":
			candidates = append(candidates, codecCandidate{"gbk", simplifiedchinese.GBK})
		}
	}

	// chardet探测失败 / 识别字符集不在映射表：fallback 默认候选队列
	if len(candidates) == 0 {
		candidates = []codecCandidate{
			{"cp1252", charmap.Windows1252},
			{"latin1", charmap.ISO8859_1},
			{"iso8859-15", charmap.ISO8859_15},
			{"iso8859-2", charmap.ISO8859_2},
			{"gbk", simplifiedchinese.GBK},
		}
	}

	return TryDecodeToUTF8Impl(raw, candidates)
}

// TryDecodeToUTF8Impl 依次遍历候选编码尝试解码
// ⚠️注意：单字节编码解码永远不会返回error，列表第一个候选直接生效
func TryDecodeToUTF8Impl(raw []byte, candidates []codecCandidate) (result string, usedCodec string) {
	for _, item := range candidates {
		outBytes, err := item.dec.NewDecoder().Bytes(raw)
		if err == nil {
			return string(outBytes), item.name
		}
	}

	// 全部编码尝试失败：兜底清洗，非法utf8替换为 � (\uFFFD)
	cleaned := cleanInvalidUTF8(string(raw))
	return cleaned, "fallback-clean"
}

// cleanInvalidUTF8 兜底清洗函数，将非法UTF8字节替换为替换字符 \uFFFD
func cleanInvalidUTF8(s string) string {
	b := []byte(s)
	var out []rune
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError && size == 1 {
			out = append(out, '\uFFFD')
			b = b[1:]
			continue
		}
		out = append(out, r)
		b = b[size:]
	}
	return string(out)
}
