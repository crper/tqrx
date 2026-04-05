<!--
这个模板的目标很简单：
1. 让 reviewer 一眼看懂这次改了什么、为什么改
2. 让验证结果和行为变更写在同一个地方
3. 如果是 TUI 改动，强制提醒附终端输出或截图
-->
## Summary

<!-- 用两三条短句写清楚“改了什么”和“为什么”，不要贴实现流水账。 -->
- What changed?
- Why is this needed?

## Validation

<!-- 勾选实际跑过的命令。没有跑过就不要勾。 -->
- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./...`
- [ ] `bash scripts/check-docs.sh`

## Notes

<!-- 这里专门放 reviewer 最容易漏掉的上下文。 -->
- Behavior changes:
- Docs updated:
- Screenshots / terminal output if the TUI changed:
