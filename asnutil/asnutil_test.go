package asnutil

import (
	"fmt"
	"strings"
	"testing"
)

/*
# 仅测试 AsnIncludeInParentAsn 性能
go test -bench="BenchmarkAsnIncludeInParentAsn" -benchmem ./asnutil
# 全量性能测试（需联网）
go test -bench=. -benchmem ./asnutil
go test -coverprofile=coverage.out ./asnutil
go tool cover -html=coverage.out
*/

func TestGetAsnOwnerByCymru1(t *testing.T) {
	r, err := GetAsnOwnerByCymru(265699)
	fmt.Println(r, err)
}

// -------------------------- AsnIncludeInParentAsn 功能测试 --------------------------

// TestAsnIncludeInParentAsn 测试 AsnIncludeInParentAsn 函数
// 覆盖：单ASN匹配、单ASN匹配段、段匹配段、边界值、无效参数等所有场景
func TestAsnIncludeInParentAsn(t *testing.T) {
	tests := []struct {
		name      string
		selfAsn   uint64
		selfMin   uint64
		selfMax   uint64
		parentAsn uint64
		parentMin uint64
		parentMax uint64
		want      bool
	}{
		// 场景1：单ASN匹配场景
		{
			name:      "自单ASN等于父单ASN",
			selfAsn:   150,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 150,
			parentMin: 0,
			parentMax: 0,
			want:      true,
		},
		{
			name:      "自单ASN不等于父单ASN",
			selfAsn:   150,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 200,
			parentMin: 0,
			parentMax: 0,
			want:      false,
		},

		// 场景2：自单ASN匹配父段场景
		{
			name:      "自单ASN在父段内（正常）",
			selfAsn:   150,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},
		{
			name:      "自单ASN等于父段最小值",
			selfAsn:   100,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},
		{
			name:      "自单ASN等于父段最大值",
			selfAsn:   200,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},
		{
			name:      "自单ASN小于父段最小值",
			selfAsn:   90,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},
		{
			name:      "自单ASN大于父段最大值",
			selfAsn:   210,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},

		// 场景3：自段匹配父段场景（核心）
		{
			name:      "自段完全在父段内（正常）",
			selfAsn:   0,
			selfMin:   150,
			selfMax:   180,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},
		{
			name:      "自段等于父段（完全重合）",
			selfAsn:   0,
			selfMin:   100,
			selfMax:   200,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},
		{
			name:      "自段部分重叠父段（右溢出）",
			selfAsn:   0,
			selfMin:   150,
			selfMax:   250,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},
		{
			name:      "自段部分重叠父段（左溢出）",
			selfAsn:   0,
			selfMin:   80,
			selfMax:   150,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},
		{
			name:      "自段完全超出父段（右侧）",
			selfAsn:   0,
			selfMin:   250,
			selfMax:   300,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},
		{
			name:      "自段完全超出父段（左侧）",
			selfAsn:   0,
			selfMin:   50,
			selfMax:   80,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},
		{
			name:      "自段包含父段（反向包含）",
			selfAsn:   0,
			selfMin:   50,
			selfMax:   250,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},
		{
			name:      "自段单点等于父段最小值",
			selfAsn:   0,
			selfMin:   100,
			selfMax:   100,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},
		{
			name:      "自段单点等于父段最大值",
			selfAsn:   0,
			selfMin:   200,
			selfMax:   200,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},

		// 场景4：混合场景（父单ASN vs 自段）
		{
			name:      "父单ASN在自段内（反向包含）",
			selfAsn:   0,
			selfMin:   100,
			selfMax:   200,
			parentAsn: 150,
			parentMin: 0,
			parentMax: 0,
			want:      false,
		},

		// 场景5：无效参数场景（边界值）
		{
			name:      "所有参数为0",
			selfAsn:   0,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 0,
			parentMax: 0,
			want:      false,
		},
		{
			name:      "自段无效（min>max）",
			selfAsn:   0,
			selfMin:   200,
			selfMax:   100,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false,
		},
		{
			name:      "父段无效（min>max）",
			selfAsn:   0,
			selfMin:   150,
			selfMax:   180,
			parentAsn: 0,
			parentMin: 200,
			parentMax: 100,
			want:      false,
		},
		{
			name:      "自段全0（仅单ASN有效）",
			selfAsn:   150,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true,
		},
		{
			name:      "父段全0（仅单ASN有效）",
			selfAsn:   150,
			selfMin:   0,
			selfMax:   0,
			parentAsn: 150,
			parentMin: 0,
			parentMax: 0,
			want:      true,
		},
		{
			name:      "自段仅max非0（无效段）",
			selfAsn:   0,
			selfMin:   0,
			selfMax:   150,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      true, // selfMin=0, selfMax=150 是有效段（0<=150），且完全在父段内
		},
		{
			name:      "自段仅min非0（无效段）",
			selfAsn:   0,
			selfMin:   150,
			selfMax:   0,
			parentAsn: 0,
			parentMin: 100,
			parentMax: 200,
			want:      false, // selfMin=150 > selfMax=0，无效段
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AsnIncludeInParentAsn(tt.selfAsn, tt.selfMin, tt.selfMax, tt.parentAsn, tt.parentMin, tt.parentMax)
			if got != tt.want {
				t.Errorf(
					"AsnIncludeInParentAsn(selfAsn=%d, selfMin=%d, selfMax=%d, parentAsn=%d, parentMin=%d, parentMax=%d) = %v, want %v",
					tt.selfAsn, tt.selfMin, tt.selfMax, tt.parentAsn, tt.parentMin, tt.parentMax, got, tt.want,
				)
			}
		})
	}
}

// -------------------------- GetAsnOwnerByCymru 功能测试 --------------------------

// TestGetAsnOwnerByCymru 测试 GetAsnOwnerByCymru 函数
// 注：网络请求测试建议用mock，此处为功能测试（可跳过网络场景）
func TestGetAsnOwnerByCymru(t *testing.T) {
	tests := []struct {
		name    string
		asn     int
		wantErr bool
		errMsg  string // 预期错误信息（包含）
	}{
		{
			name:    "ASN为0（无效）",
			asn:     0,
			wantErr: true,
			errMsg:  "invalid (must be positive integer)",
		},
		{
			name:    "ASN为负数（无效）",
			asn:     -123,
			wantErr: true,
			errMsg:  "invalid (must be positive integer)",
		},
		{
			name:    "ASN为有效正数（网络测试，需联网）",
			asn:     266087,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "ASN不存在（网络测试，需联网）",
			asn:     9999999,
			wantErr: false, // 无结果但不返回错误
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetAsnOwnerByCymru(tt.asn)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAsnOwnerByCymru(%d) error = %v, wantErr %v", tt.asn, err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("GetAsnOwnerByCymru(%d) error message = %v, want contain %q", tt.asn, err, tt.errMsg)
			}
		})
	}
}

// -------------------------- 性能测试 --------------------------

// BenchmarkAsnIncludeInParentAsn 性能测试 AsnIncludeInParentAsn 函数
func BenchmarkAsnIncludeInParentAsn(b *testing.B) {
	// 测试场景：自段完全在父段内（高频场景）
	selfAsn := uint64(0)
	selfMin := uint64(150)
	selfMax := uint64(180)
	parentAsn := uint64(0)
	parentMin := uint64(100)
	parentMax := uint64(200)

	// 预热
	_ = AsnIncludeInParentAsn(selfAsn, selfMin, selfMax, parentAsn, parentMin, parentMax)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AsnIncludeInParentAsn(selfAsn, selfMin, selfMax, parentAsn, parentMin, parentMax)
	}
}

// BenchmarkAsnIncludeInParentAsn_SingleMatch 性能测试：单ASN匹配场景
func BenchmarkAsnIncludeInParentAsn_SingleMatch(b *testing.B) {
	selfAsn := uint64(150)
	selfMin := uint64(0)
	selfMax := uint64(0)
	parentAsn := uint64(150)
	parentMin := uint64(0)
	parentMax := uint64(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AsnIncludeInParentAsn(selfAsn, selfMin, selfMax, parentAsn, parentMin, parentMax)
	}
}

// BenchmarkAsnIncludeInParentAsn_NoMatch 性能测试：无匹配场景
func BenchmarkAsnIncludeInParentAsn_NoMatch(b *testing.B) {
	selfAsn := uint64(0)
	selfMin := uint64(250)
	selfMax := uint64(300)
	parentAsn := uint64(0)
	parentMin := uint64(100)
	parentMax := uint64(200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AsnIncludeInParentAsn(selfAsn, selfMin, selfMax, parentAsn, parentMin, parentMax)
	}
}

// BenchmarkGetAsnOwnerByCymru 性能测试 GetAsnOwnerByCymru 函数（需联网）
func BenchmarkGetAsnOwnerByCymru(b *testing.B) {
	asn := 266087 // 有效ASN

	// 预热（仅一次，避免重复网络请求）
	_, _ = GetAsnOwnerByCymru(asn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetAsnOwnerByCymru(asn)
	}
}
