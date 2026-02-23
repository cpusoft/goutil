package asnutil

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"context"

	"github.com/cpusoft/goutil/belogs"
)

// AsnIncludeInParentAsn 判断自ASN（或ASN段）是否被父ASN（或ASN段）完全包含
// 核心规则：
// 1. 自单ASN == 父单ASN → true
// 2. 自单ASN 落在 父ASN段（parentMin-parentMax）内 → true
// 3. 自ASN段（selfMin-selfMax）完全落在 父ASN段（parentMin-parentMax）内 → true
// 4. 部分包含/重叠/父ASN在自段内 → false
// selfAsn: 单个ASN值；selfMin/selfMax: 自ASN段的起止（需满足selfMin<=selfMax才视为有效段）
// parentAsn: 父单个ASN值；parentMin/parentMax: 父ASN段的起止（需满足parentMin<=parentMax才视为有效段）
func AsnIncludeInParentAsn(selfAsn, selfMin, selfMax, parentAsn, parentMin, parentMax uint64) bool {
	// 定义「有效单ASN」和「有效ASN段」的规则
	isSelfSingleValid := selfAsn != 0                                                  // 自单ASN有效
	isSelfRangeValid := selfMin <= selfMax && (selfMin != 0 || selfMax != 0)           // 自段有效（min<=max且非全0）
	isParentSingleValid := parentAsn != 0                                              // 父单ASN有效
	isParentRangeValid := parentMin <= parentMax && (parentMin != 0 || parentMax != 0) // 父段有效

	// 任意一方无有效ASN/段 → 返回false
	if !isSelfSingleValid && !isSelfRangeValid {
		return false
	}
	if !isParentSingleValid && !isParentRangeValid {
		return false
	}

	// 场景1：自单ASN == 父单ASN → true
	if isSelfSingleValid && isParentSingleValid && selfAsn == parentAsn {
		return true
	}

	// 场景2：自单ASN 落在 父ASN段内 → true
	if isSelfSingleValid && isParentRangeValid {
		if parentMin <= selfAsn && selfAsn <= parentMax {
			return true
		}
	}

	// 场景3：自ASN段 完全落在 父ASN段内 → true（核心：仅完全包含才返回true）
	if isSelfRangeValid && isParentRangeValid {
		if parentMin <= selfMin && selfMax <= parentMax {
			return true
		}
	}

	// 所有其他场景（部分重叠、父ASN在自段内、自段超出父段等）→ false
	return false
}

// GetAsnOwnerByCymru 保持之前修复后的逻辑（无修改）
func GetAsnOwnerByCymru(asn int) (string, error) {
	// 1. 入参校验：ASN为非负整数
	if asn <= 0 {
		return "", errors.New("asn is invalid (must be positive integer)")
	}

	lookupUrl := fmt.Sprintf(`AS%d.asn.cymru.com`, asn)
	belogs.Debug("GetAsnOwnerByCymru():asn lookupUrl:", asn, lookupUrl)

	// 2. 带超时的DNS查询（5秒超时，避免挂起）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var txt []string
	var err error
	// 自定义DNS解析器，带超时控制
	resolver := &net.Resolver{}
	if txt, err = resolver.LookupTXT(ctx, lookupUrl); err != nil {
		belogs.Error("GetAsnOwnerByCymru(): lookupUrl fail:", lookupUrl, asn, err)
		return "", fmt.Errorf("dns lookup failed: %w", err) // 包装错误，保留原始错误链
	}

	// 3. 空结果处理：明确返回空字符串（无错误）
	if len(txt) == 0 || strings.TrimSpace(txt[0]) == "" {
		belogs.Info("GetAsnOwnerByCymru(): lookupUrl txt is empty:", asn, txt)
		return "", nil
	}

	// 4. 拆分TXT记录，严格边界校验
	// 示例格式："266087 | BR | lacnic | 2017-03-13 | Orbitel Telecomunicacoes e Informatica Ltda, BR"
	splits := strings.Split(txt[0], "|")
	// 过滤空元素（处理多余的|或空格）
	cleanSplits := make([]string, 0, len(splits))
	for _, s := range splits {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			cleanSplits = append(cleanSplits, trimmed)
		}
	}

	// 校验有效字段数（至少需要3个字段：ASN | 国家 | 注册机构）
	if len(cleanSplits) < 3 {
		belogs.Error("GetAsnOwnerByCymru(): Split fail (insufficient fields):", asn, txt, splits)
		return "", fmt.Errorf("invalid txt record format: %s", txt[0])
	}

	// 5. 安全拼接结果，避免索引越界
	org := cleanSplits[len(cleanSplits)-1] // 机构名称
	registry := cleanSplits[2]             // 注册机构（lacnic/arin等）
	result := fmt.Sprintf("%s,%s", org, registry)

	belogs.Debug("GetAsnOwnerByCymru(): lookup success:", asn, result)
	return result, nil
}
