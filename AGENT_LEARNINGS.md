# AGENT_LEARNINGS

- 2026-04-03: 在 Codex 沙箱里跑 Go 校验时，默认 `GOCACHE` 可能因为写 `~/Library/Caches/go-build` 被拒。稳定做法是加 `GOCACHE=/tmp/go-build`。
- 2026-04-03: 这个仓库当前离线可用的依赖在默认模块缓存里；不要把 `GOMODCACHE` 切到 `/tmp`，否则会触发重新下载并因网络受限失败。离线校验优先加 `GOPROXY=off`。
- 2026-04-03: `GOCACHE=/tmp/go-build GOPROXY=off go build ./...` 在当前环境可能打印模块 stat cache 写入警告，但只要退出码是 `0`，构建本身仍然是成功的。
