Params / Values / Files 重构性能对照（2026-09-24）

本轮结果表明：Params/Values 的构建、更新和并发 Set/Get 成本下降；读锁和深拷贝也引入了开销，其中 Files.Get 的成本增加最明显。编码耗时波动较大，不能据此推断实际 HTTP 请求吞吐量。

基线为 `21e90ca`（修复请求边界问题并补充回归测试）；新版为本次尚未提交的锁与深拷贝重构。基线已包含上次文件死锁修复，因此本次仅衡量后续重构的差异。

环境：Windows/amd64，Intel Core Ultra 7 155H，Go 1.26.2。使用相同的[基准代码](../../url/collections_bench_test.go)，分别编译两个测试程序；`GOMAXPROCS=4`、每项 `200ms`、每版 6 个样本。两版共交替运行 6 轮：奇数轮先旧后新，偶数轮先新后旧。计时程序串行运行，期间未运行编译或测试任务；未固定 CPU 核心、频率或操作系统后台负载。

两份基准源文件 SHA-256 相同：`88BE67728BA32449463B26D8A85FFC66710F6C780677FE2D068852039F158C35`。

测试场景：

- `Build16`：创建集合，依次添加 16 个键，每键 2 个值；值为 `first value` 和 `second&value`。
- `SetExisting` / `Get`：在预先构造的集合中循环访问 16 个键；更新将该键替换为一个值。
- `Snapshot`：调用 Params/Values 的 `Values()`；旧版只复制 map，新版连同值切片一起复制，安全保证不同。
- `Encode` / `EncodeFields`：编码完整的 16 键、32 值工作集；Files 使用普通 multipart 字段。
- `EncodeFile64KiB`：读取并编码一个本地 64 KiB 文件，包含文件打开和读取，文件通常已进入系统缓存；不包含网络上传。
- `ParallelSetGet`：4 个并发 worker 共享预先构造的 16 键集合，每次操作包含一次 Set 和一次 Get。旧版在此固定键场景可安全运行。

每次迭代的工作量有界，不使用随 b.N 增长的无限 Add。旧版 Params/Values 并发 Add 存在丢值和竞态，因此没有将它作为正确实现的吞吐量基线。新增的 Clone 没有旧版同名接口，本表未单独比较 Clone，也未测量 Session 合并、请求准备或完整网络请求。

以下均为 6 个样本的中位数。变化 =（新版耗时 / 旧版耗时 − 1）× 100%；负值表示耗时下降。`B/op` 为每次操作的分配字节数，`allocs/op` 为分配次数。

| 场景 | 旧版 ns/op | 新版 ns/op | 耗时变化 | B/op：旧 → 新 | allocs/op：旧 → 新 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Params/Build16 | 5377 | 3409.5 | -36.6% | 5098 → 3912 | 140 → 45 |
| Params/SetExisting | 143.85 | 56.13 | -61.0% | 104 → 16 | 4 → 1 |
| Params/Get | 13.36 | 19.02 | +42.4% | 0 → 0 | 0 → 0 |
| Params/Snapshot | 1217 | 1316.5 | +8.2% | 2584 → 2008 | 7 → 20 |
| Params/Encode | 1885.5 | 2148.5 | +13.9% | 2272 → 2272 | 39 → 39 |
| Params/ParallelSetGet | 234.6 | 131.8 | -43.8% | 104 → 16 | 4 → 1 |
| Values/Build16 | 5876.5 | 3401 | -42.1% | 5097 → 3912 | 140 → 45 |
| Values/SetExisting | 142.85 | 56.86 | -60.2% | 104 → 16 | 4 → 1 |
| Values/Get | 16.13 | 19.45 | +20.6% | 0 → 0 | 0 → 0 |
| Values/Snapshot | 2177 | 1659.5 | -23.8% | 2584 → 2008 | 7 → 20 |
| Values/Encode | 13919.5 | 13113.5 | -5.8% | 3648 → 3648 | 149 → 149 |
| Values/ParallelSetGet | 317.45 | 222.85 | -29.8% | 104 → 16 | 4 → 1 |
| Files/Build16 | 15119 | 12808 | -15.3% | 18368 → 14280 | 140 → 109 |
| Files/Get | 61.61 | 324.85 | +427.3% | 0 → 336 | 0 → 2 |
| Files/EncodeFields | 50591.5 | 59499 | +17.6% | 33280.5 → 46108 | 494 → 580 |
| Files/EncodeFile64KiB | 137802.5 | 148835.5 | +8.0% | 171325.5 → 172150.5 | 48 → 55 |
| Files/ParallelSetGet | 831.7 | 956.8 | +15.0% | 744 → 680 | 5 → 5 |

可以较明确看到的收益是 Params/Values 更新已有键的耗时下降约 60%，分配从 104 B、4 次降至 16 B、1 次；并发 Set/Get 在全部 6 轮中都更快，中位数下降约 30%～44%。构建集合的分配次数从 140 次降至 45 次。移除 sync.Map 的装箱以及键的线性查找，有助于降低这些路径的成本。

代价主要体现在读取和快照：Params.Get 的中位数增加约 5.66 ns；Values.Get 的中位数增加约 3.33 ns，但样本存在重叠。Files.Get 从返回内部 map 改为返回独立副本，中位数由 61.61 ns 增至 324.85 ns（约 5.27 倍），新增 336 B、2 次分配，调用方修改返回值不再改变集合。

Params/Values 的 Snapshot 分配次数从 7 次升至 20 次，因为现在复制 16 个值切片；预分配 map 又使总分配字节数从 2584 B 降至 2008 B。两类几乎相同的快照实现测得相反的耗时变化，结合下表的波动，不能据此宣称一种快照一定更快。它们的深拷贝隔离能力是确定的，精确耗时幅度仍受环境影响。

Files.Encode 先复制元数据，再释放集合锁执行 multipart 编码与文件读取，避免在磁盘 I/O 期间阻塞修改操作。32 个普通字段的分配量约增加 38.5%；单文件编码增加约 824 B、7 次分配。该测试没有直接测量缩短持锁时间对并发写入延迟的收益。

编码中位数变化为 Params +13.9%、Values −5.8%、Files 普通字段 +17.6%、64 KiB 文件 +8.0%。Params/Values 编码的分配量均未变化，不能将其耗时变化归因于新增深拷贝。特别是文件编码，新版单轮最大值达到 314920 ns，明显高于其余样本；这些编码数字仅描述本次测量，不作为稳定加速或退化幅度的承诺。

以下列出全部场景的样本范围与配对方向。范围是实测最小值至最大值，不是置信区间；“新版更快轮数”逐轮比较两个版本，不能代替统计显著性检验。

| 场景 | 旧版 ns/op 范围 | 新版 ns/op 范围 | 新版更快轮数 |
| --- | ---: | ---: | ---: |
| Params/Build16 | 4564～7656 | 3162～6156 | 5/6 |
| Params/SetExisting | 129.1～155.4 | 52.26～73 | 6/6 |
| Params/Get | 11.86～14.58 | 16.95～25.5 | 0/6 |
| Params/Snapshot | 1046～1814 | 1081～2382 | 2/6 |
| Params/Encode | 1792～2332 | 1802～3903 | 2/6 |
| Params/ParallelSetGet | 225.3～251 | 117～179.3 | 6/6 |
| Values/Build16 | 4940～6954 | 3146～4499 | 6/6 |
| Values/SetExisting | 134.2～166.3 | 49.3～73.77 | 6/6 |
| Values/Get | 12.03～24.34 | 16.45～25.94 | 2/6 |
| Values/Snapshot | 1641～2304 | 1057～2144 | 5/6 |
| Values/Encode | 12926～16174 | 11557～14024 | 6/6 |
| Values/ParallelSetGet | 307.1～474.7 | 201.2～229.3 | 6/6 |
| Files/Build16 | 12162～17934 | 11651～15443 | 6/6 |
| Files/Get | 59.49～76.86 | 282.9～353.6 | 0/6 |
| Files/EncodeFields | 40661～64240 | 53524～94334 | 1/6 |
| Files/EncodeFile64KiB | 124642～147298 | 124176～314920 | 2/6 |
| Files/ParallelSetGet | 723.8～871 | 930.7～1940 | 0/6 |

原始结果：[before.txt](before.txt)、[after.txt](after.txt)。各文件内的 6 段结果依次对应第 1～6 轮。本报告以这组固定工作集、交替运行的数据为准，不采用早先无限增长 Add 或单次测量的结论。

可在仓库根目录使用 PowerShell 复现。先确保当前工作区包含本次重构；临时工作树仅加入同一份 benchmark 文件。输出写入唯一临时目录，不覆盖本次记录：

```powershell
$ErrorActionPreference = 'Stop'
$repoDir = (Get-Location).Path
$runDir = Join-Path $env:TEMP ('requests-collections-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $runDir | Out-Null
$beforeDir = Join-Path $runDir 'before'
git worktree add --detach $beforeDir 21e90ca
if ($LASTEXITCODE -ne 0) { throw '创建基线工作树失败' }
Copy-Item -LiteralPath (Join-Path $repoDir 'url/collections_bench_test.go') -Destination (Join-Path $beforeDir 'url/collections_bench_test.go')
$beforeExe = Join-Path $runDir 'before.test.exe'
$afterExe = Join-Path $runDir 'after.test.exe'
Push-Location $beforeDir
try {
    go test -c -o $beforeExe ./url
    if ($LASTEXITCODE -ne 0) { throw '编译基线失败' }
} finally { Pop-Location }
go test -c -o $afterExe ./url
if ($LASTEXITCODE -ne 0) { throw '编译新版失败' }
$programs = @{ before = $beforeExe; after = $afterExe }
foreach ($round in 1..6) {
    $order = if ($round % 2 -eq 1) { @('before', 'after') } else { @('after', 'before') }
    foreach ($version in $order) {
        $exe = $programs[$version]
        $output = & $exe '-test.run=^$' '-test.bench=^BenchmarkCollections' '-test.benchmem' '-test.benchtime=200ms' '-test.cpu=4' '-test.count=1'
        if ($LASTEXITCODE -ne 0) { throw "基准运行失败：$version，第 $round 轮" }
        $output | Out-File -Encoding UTF8 -Append (Join-Path $runDir "$version.txt")
    }
}
# 删除本脚本复制的文件后，用 Git 移除临时工作树；保留测试程序和结果。
Remove-Item -LiteralPath (Join-Path $beforeDir 'url/collections_bench_test.go')
git worktree remove $beforeDir
if ($LASTEXITCODE -ne 0) { throw '清理基线工作树失败' }
Write-Output "结果目录：$runDir"
```

正确性验证已通过：

```text
go test . ./models ./url ./transport ./utils ./libs ./libs_dist
go test -race . ./models ./url ./transport ./utils ./libs ./libs_dist
```

新增回归覆盖同键并发 Add、修改与快照并发、快照隔离、零值可用、键顺序、multipart 文件与重复字段、共享 Files 的顺序/并发复用，以及 Session 参数合并隔离。项目现有 `examples` 目录在同一个包内定义多个 main，导致 `go test ./...` 仍无法整体通过；上述命令显式覆盖可构建的库包。
