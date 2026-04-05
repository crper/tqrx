package tui

import (
	"testing"

	"github.com/crper/tqrx/internal/core"
	"github.com/crper/tqrx/internal/render"
)

func BenchmarkCollectLevelModules(b *testing.B) {
	benchmarks := []struct {
		name    string
		content string
	}{
		{
			name:    "sparse",
			content: "https://example.com/benchmark/level-modules",
		},
		{
			name:    "dense",
			content: "发的上课了放假快乐 sd 卡放假啦 sd 卡房间为埃及人 weakly 放假 ADSL 客服啊圣诞快乐发阿斯蒂芬看阿斯蒂芬阿斯蒂芬跨时代开放啦是的副卡就是的罚款了是的积分啊上看到了放假啊圣诞快乐发阿斯蒂芬克拉斯都发开始了地方啊圣诞快乐发阿斯蒂芬开了撒旦法是短发收到了客服阿斯蒂芬集卡老师的飞机 adslkffafffffffffffffffffjjlkje2kjeio2u09u 批发第三方届奥斯卡放假啊放假啊酸辣粉卡时间发快手发阿斯蒂芬撒旦法撒旦法是",
		},
	}

	for _, tt := range benchmarks {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				modules := collectLevelModules(tt.content)
				if len(modules) == 0 {
					b.Fatal("collectLevelModules() = empty map")
				}
			}
		})
	}
}

func BenchmarkLevelModulesForContent(b *testing.B) {
	content := "https://example.com/benchmark/level-modules/cache"
	missContent := content + "/miss"
	cached := collectLevelModules(content)
	if len(cached) == 0 {
		b.Fatal("collectLevelModules() warm-up = empty map")
	}

	model := NewModel(nil)
	model.levelModulesContent = content
	model.levelModules = cached

	b.Run("cached", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			modules := model.levelModulesForContent(content)
			if len(modules) != len(cached) {
				b.Fatalf("levelModulesForContent() len = %d, want %d", len(modules), len(cached))
			}
		}
	})

	b.Run("miss", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			modules := model.levelModulesForContent(missContent)
			if len(modules) == 0 {
				b.Fatal("levelModulesForContent() miss = empty map")
			}
		}
	})
}

func BenchmarkPreviewMetaLineCount(b *testing.B) {
	model := NewModel(nil)
	req, err := core.Normalize(core.Request{
		Content: "fasdkfjasdlkfasdkfldajsfklajflkfaksd\nlfjadsklfjadsklfjadsklf\njadsklfjadsklfjadsklfjadsklfja\ndsfasdfhkjhkjjhkjhjhk",
		Level:   string(core.LevelQuart),
		Source:  core.SourceTUI,
	})
	if err != nil {
		b.Fatalf("Normalize() error = %v", err)
	}
	prepared, err := render.NewEngine().Prepare(req)
	if err != nil {
		b.Fatalf("Prepare() error = %v", err)
	}

	model.prepared = prepared
	model.format = req.Format
	model.level = req.Level
	model.size.SetValue("256")
	model.output.SetValue("/tmp/二维码导出/这是一个特别特别长的输出路径-用于验证预览元信息换行是否稳定并且和布局估算一致.png")
	model.levelModules = collectLevelModules(req.Content)
	model.levelModulesContent = req.Content
	model.preview.SetWidth(56)
	model.preview.SetHeight(20)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lines := model.previewMetaLineCount(69)
		if lines <= 0 {
			b.Fatalf("previewMetaLineCount() = %d, want positive line count", lines)
		}
	}
}
