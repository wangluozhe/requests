# Transport 基准

基准命令：

```text
go test -run '^$' -bench '^(BenchmarkSettingsSnapshotTLS|BenchmarkSettingsSnapshotHTTP2|BenchmarkTransportCacheLookup|BenchmarkTransportCacheHitWithSnapshot|BenchmarkTransportCacheHitWithoutSnapshot|BenchmarkTLSSettingsHash|BenchmarkHTTP2SettingsHash)$' -benchtime=200ms -count=6 -cpu=4
```

基线为提交 `6746a5c`，前后测试均在 Intel(R) Core(TM) Ultra 7 155H、Windows amd64、Go 1.26.2 上执行。下表取 6 次结果的中位数：

| 基准 | 对照路径 | 优化后 | 相对耗时（优化后/对照路径） |
| --- | ---: | ---: | ---: |
| TLS 快照 | 旧版直接返回指针：1.896 ns/op，0 B，0 allocs | 不可变深拷贝：约 109 ns/op，216 B，5 allocs | 约 57 倍（语义成本） |
| HTTP/2 快照 | 旧版直接返回指针：1.750 ns/op，0 B，0 allocs | 不可变深拷贝：约 127 ns/op，128 B，4 allocs | 约 73 倍（语义成本） |
| TLS 字段哈希 | JSON：约 763 ns/op，376 B，2 allocs | 字段哈希：约 271 ns/op，24 B，1 alloc | 约降低 64% |
| HTTP/2 字段哈希 | JSON：约 829 ns/op，376 B，2 allocs | 字段哈希：约 379 ns/op，24 B，1 alloc | 约降低 54% |
| Transport 缓存查找 | 旧版 map 查找：13.985 ns/op，0 B，0 allocs | LRU 查找：约 25 ns/op，0 B，0 allocs | 单独查找有噪声，见缓存命中项 |
| 缓存命中配置路径 | JSON 哈希+快照：约 2007.5 ns/op，1096 B，13 allocs | 字段哈希、命中时不复制：约 646 ns/op，48 B，2 allocs | 约降低 68% |

TLS/HTTP/2 的旧版快照基准只是返回原始指针，没有复制成本，因此不能把它当作同语义的性能基线。不可变快照必然需要复制；当前实现使用字段级复制，并把复制延迟到缓存未命中时，避免缓存命中请求重复分配。字段哈希跳过 JSON 序列化，缓存命中路径相较之前的 JSON 哈希和提前复制约快 78%。原始结果见 `before.txt`、`after.txt`、`field-hash.txt`。
