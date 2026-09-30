# territory-capture

领地占领：按在线人数与真实时间推进进度，无人争夺时衰减，过阈值翻转归属，只用 Go 标准库。

```
go test ./...
go run ./cmd/territory-capture --sample time
go run ./cmd/territory-capture --sample decay
go run ./cmd/territory-capture --sample flip
```

## 口径（README 为准）

- **按真实时间推进**：本次增量等于净人数乘 `RatePerSec` 乘 `dtMs/1000`，与调用次数无关；
  同样一人争夺 1 秒与 10 秒的增量必须差 10 倍。
- **无人争夺要衰减**：双方都是 0 时按 `DecayPerSec` 乘 `dtMs/1000` 向 0 回落。
- **过阈值翻转**：进度到 `+FlipThreshold` 归 a、到 `-FlipThreshold` 归 b，翻转后进度清零并把
  `Flips` 加一；进度始终夹在 `[-100, 100]`。
- **查询代价**：`Snapshot` 不许每次全表扫描，`Scanned` 不随点数线性增长。
- **规模**：单场 5000 个点、每秒 10 万次推进与查询，内存 O(点数)。

## 输出契约（不改格式）

```
one=10 ten=100
progress=20
progress=0 owner=a flips=1
scanned<=3000
```

## 目录

```
territory.go             占领进度与翻转
cmd/territory-capture    命令行入口
territory_test.go        go test 用例
```
