# gameops v1.0.0 — 生产发布说明

单一 Go 二进制承载游戏长线运营引擎全部能力。**静态链接（CGO_ENABLED=0）、零外部依赖、精确有理数计算、支持所有游戏品类。**

## 发布产物

`dist/v1.0.0/`：

| 平台 | 路径 |
|---|---|
| Windows x64 | `gameops-windows-amd64.exe` |
| Windows ARM64 | `gameops-windows-arm64.exe` |
| Linux x64 | `gameops-linux-amd64` |
| Linux ARM64 | `gameops-linux-arm64` |
| macOS x64 | `gameops-darwin-amd64` |
| macOS ARM64 (Apple Silicon) | `gameops-darwin-arm64` |

GitHub Release: https://github.com/neko233-com/game-ops/releases/tag/v1.0.0

元数据：`version.json`（版本/commit/构建时间/Go 版本/目标列表）、`SHA256SUMS`（文件名与 Release asset 一致，可直接 `sha256sum -c`）。

## 校验完整性

```bash
cd dist/v1.0.0 && sha256sum -c SHA256SUMS      # macOS: shasum -a 256 -c SHA256SUMS
```

```powershell
cd dist/v1.0.0
Get-Content SHA256SUMS | ForEach-Object {
  $h,$p = $_ -split '  '
  $actual = (Get-FileHash -Algorithm SHA256 $p).Hash.ToLower()
  if ($h -ne $actual) { throw "MISMATCH: $p" }
}
"all OK"
```

## 安装

```powershell
# Windows: 安装到 ~/bin（也可 -InstallDir 指定）
powershell -ExecutionPolicy Bypass -File scripts/build-release.ps1 -Install
# 将 ~/bin 加入 PATH，之后任意目录可运行 gameops
```

```bash
# macOS/Linux: 复制到 PATH 目录
cp dist/v1.0.0/gameops-$(uname -s | tr A-Z a-z)-amd64 /usr/local/bin/gameops
chmod +x /usr/local/bin/gameops
```

源码构建：`bash scripts/build-release.sh`（或 PowerShell 版），可设 `VERSION=x.y.z`。

## 用法

```bash
gameops version                                    # 版本/commit/构建时间/Go 版本
gameops help                                       # 15 个命令清单
gameops boss brief gacha -values examples/gacha-warn.json
gameops boss decide gacha -values examples/gacha-warn.json -html cockpit.html
gameops opsreport examples/gacha.json
gameops numcalc '1/3 + 2/7'
gameops catalog derive cozy-farming -tags cozy,farming
```

15 个子命令：`boss` `catalog` `opsreport` `econcheck` `normalize` `numgen` `curve` `numcalc` `formula` `genre` `strategist` `judge` `ingest` `proanalysis` `htmlreport`。

## 生产级保证

- **单一二进制**：所有能力在一个可执行文件；子命令逻辑直接复用原 `cmd/*` 包（`Main()`），零重复代码、零行为差异。
- **静态链接**：`CGO_ENABLED=0` + `-trimpath`，无动态库依赖，可进最小容器。
- **版本可追溯**：`-ldflags` 注入 version/commit/date，`gameops version` 可查；`version.json` 记录构建元数据。
- **确定性**：并行摄取/敏感度分析在固定输入下输出逐字节一致（5 次重复哈希相同）。
- **完整性**：6 平台 SHA256SUMS，`sha256sum -c` 全 OK。
- **错误语义**：未知命令 exit 1；指数公式守卫 exit 1；panic 恢复 exit 70（sysexits EX_SOFTWARE）。
- **测试**：`go build` / `go vet` / `go test -count=1 ./...` 全绿；20+ 测试覆盖引擎、开放目录与 boss 决策层。

注：`go test -race` 需 CGO 与 C 编译器，本机无 gcc 故未运行；改用高并行重复执行的确定性校验替代。

## 架构

```
cmd/gameops/main.go          单一入口：分发 + 版本 + panic 恢复
cmd/<name>/main.go           薄启动器（package main），保留 go run ./cmd/<name>
internal/commands/<name>/    命令实现（package <name>cmd），与 gameops 共享零重复
internal/buildinfo           版本元数据
boss/                        决策翻译层
genre/catalog.go        数据驱动品类（Derive / Upsert / GetAny / JSON spec）
ops/ops.go              EvaluateAny（未知品类自动派生）
factory/universal.go    通用兜底策略
```

严格 API（未知报错）与宽松 API（自动派生）双轨并存：引擎不破坏既有语义，boss/catalog 走宽松轨实现"支持所有"。
