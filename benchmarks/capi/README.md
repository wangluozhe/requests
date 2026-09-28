# C API 生命周期与 Cookie 基准

基准命令：

```text
go test -run '^$' -bench '^BenchmarkCSession' -benchmem -benchtime=200ms -count=6 -cpu=4 ./libs_dist
```

测试环境：Windows amd64，Intel(R) Core(TM) Ultra 7 155H，Go 1.26.2。
结果使用 6 次样本的中位数，基准只测 Session 查找、Cookie Jar 初始化和生命周期管理，不包含网络传输。

| 场景 | 重构前 | 重构后 | 变化 |
| --- | ---: | ---: | ---: |
| 请求获取 Session（每次创建 Jar 并查池 -> acquire/release） | 50.84 ns/op，96 B，2 allocs | 40.83 ns/op，0 B，0 allocs | 约下降 19.7% |

重构前基准模拟旧 `GetSession` 的命中路径：每次创建并替换 Cookie Jar，再从池中查找 Session。重构后基准覆盖请求进入和退出时的完整 `acquireSession` / `releaseSession` 生命周期。新路径每次减少约 10 ns，并从 96 B、2 次分配降为零分配；此外，Cookie Jar 在同一 Session ID 的请求间保留，因此响应设置的 Cookie 能在后续请求发送。

裸 `sync.Map.Load` 基线为 8.96 ns/op；完整引用计数路径为约 40.83 ns/op，额外成本用于阻止 `freeSession` 与活动请求并发关闭。`BenchmarkCSessionLookupAfter` 单独测得 Session 池查找约 15.89 ns/op、零分配。以上是代码路径微基准，不包括网络传输。

完整原始输出见 `before.txt` 和 `after.txt`。
