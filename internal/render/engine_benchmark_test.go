package render

import (
	"testing"

	"github.com/crper/tqrx/internal/core"
)

func BenchmarkPrepare(b *testing.B) {
	req := mustNormalizeBenchmarkRequest(b, core.Request{
		Content: "https://example.com/benchmark/prepare?payload=abcdefghijklmnopqrstuvwxyz0123456789",
		Level:   string(core.LevelHigh),
		Size:    "320",
		Source:  core.SourceCLIArg,
	})

	b.Run("cold", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			engine := NewEngine()
			prepared, err := engine.Prepare(req)
			if err != nil {
				b.Fatalf("Prepare() error = %v", err)
			}
			if prepared == nil {
				b.Fatal("Prepare() = nil, want prepared value")
			}
		}
	})

	b.Run("cached", func(b *testing.B) {
		engine := NewEngine()
		prepared, err := engine.Prepare(req)
		if err != nil {
			b.Fatalf("Prepare() warm-up error = %v", err)
		}
		if prepared == nil {
			b.Fatal("Prepare() warm-up = nil, want prepared value")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			prepared, err := engine.Prepare(req)
			if err != nil {
				b.Fatalf("Prepare() error = %v", err)
			}
			if prepared == nil {
				b.Fatal("Prepare() = nil, want prepared value")
			}
		}
	})
}

func BenchmarkPreviewFit(b *testing.B) {
	sparse := mustPreparedBenchmarkQR(b, core.Request{
		Content: "preview-fit-sparse",
		Source:  core.SourceCLIArg,
	})
	dense := mustPreparedBenchmarkQR(b, core.Request{
		Content: "发的上课了放假快乐 sd 卡放假啦 sd 卡房间为埃及人 weakly 放假 ADSL 客服啊圣诞快乐发阿斯蒂芬看阿斯蒂芬阿斯蒂芬跨时代开放啦是的副卡就是的罚款了是的积分啊上看到了放假啊圣诞快乐发阿斯蒂芬克拉斯都发开始了地方啊圣诞快乐发阿斯蒂芬开了撒旦法是短发收到了客服阿斯蒂芬集卡老师的飞机 adslkffafffffffffffffffffjjlkje2kjeio2u09u 批发第三方届奥斯卡放假啊放假啊酸辣粉卡时间发快手发阿斯蒂芬撒旦法撒旦法是",
		Level:   string(core.LevelHigh),
		Source:  core.SourceCLIArg,
	})

	benchmarks := []struct {
		name     string
		prepared *Prepared
		width    int
		height   int
	}{
		{
			name:     "sparse_large_viewport",
			prepared: sparse,
			width:    60,
			height:   24,
		},
		{
			name:     "dense_small_viewport",
			prepared: dense,
			width:    24,
			height:   10,
		},
	}

	for _, tt := range benchmarks {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				preview := tt.prepared.PreviewFit(tt.width, tt.height)
				if preview == "" {
					b.Fatal("PreviewFit() = empty string")
				}
			}
		})
	}
}

func mustNormalizeBenchmarkRequest(b *testing.B, req core.Request) core.NormalizedRequest {
	b.Helper()

	normalized, err := core.Normalize(req)
	if err != nil {
		b.Fatalf("Normalize() error = %v", err)
	}
	return normalized
}

func mustPreparedBenchmarkQR(b *testing.B, req core.Request) *Prepared {
	b.Helper()

	prepared, err := NewEngine().Prepare(mustNormalizeBenchmarkRequest(b, req))
	if err != nil {
		b.Fatalf("Prepare() error = %v", err)
	}
	return prepared
}
