# voice-duck

语音闪避：某条总线上有人说话时，压低该总线上的其它音源。只用 Go 标准库。

```
go test ./...
go run ./cmd/duck --sample=cross-bus
go run ./cmd/duck --sample=two-talkers
go run ./cmd/duck --sample=exempt
go run ./cmd/duck --sample=release
go run ./cmd/duck --sample=work
```

不带 `--sample` 时从标准输入读 JSONL 事件，字段为 `id`、`bus`、`speaker`、`action`、`exempt`、`at`。

## 口径

- **说话状态按总线计**：每条总线各自维护「当前有几个人在说话」的计数，总线之间互不影响。
- **说话计数**：`start` 加一，`end` 减一；计数大于零才算有人说话。同一说话人重复 `start` 不重复计数。
- **闪避**：本总线有人说话时，落在本总线上的非豁免音源增益为 `Threshold`，否则为 `1`。
- **豁免覆盖层**：`exempt` 为真的音源不受闪避影响。
- **释放窗口**：本总线最后一个人停止说话后 `ReleaseMs` 毫秒内仍按闪避处理，超过才恢复。
- **幂等**：同一个 `eventID` 只处理一次。

## 不变量

- 别处有人说话不改变本总线的增益。
- 说话计数不会降到零以下。
- 判定按总线索引进行，`Scanned` 不得随已知总线总数放大。

## 输出契约

每条输入打印一行：

```
{"id": <原样>, "gain": <增益>, "ducked": <是否压低>}
```

`--sample=work` 只打印一行汇总：`{"events": N, "scanned": N}`。
