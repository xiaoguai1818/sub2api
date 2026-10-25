# 2026-10-25 第四批独立回移台账：计费和断流

## 独立基线与边界

- 目标仓库：`CherryLover/sub2api`
- 基线：`main@6af365be55481a5a14d86222eefc8e742b70d738`（v0.1.206）。
- 独立分支：`backport/2026-10-25-batch4-billing`，直接从上述 main 建立，不合并或依赖前三批未合并 PR #32、#33、#34。不得自动合并/发布。
- 严格限制在精简版已有的网关、Usage、图片及分组计费路径；保留现有安全回退与计费策略。不恢复支付/营销/站点登录等模块。
- 基于 upstream PR head 校对行为，不能直接机械 cherry-pick；重叠逻辑保留精简版独立修复。安全扫描要求独立自包含：Go 1.26.10、DOMPurify 3.4.17、follow-redirects 1.16.1。

## B01–B08 实施记录

| ID | 上游参考 | 精简版独立适配 | 验收重点 |
|---|---|---|---|
| B01 | [#8168](https://github.com/Wei-Shaw/sub2api/pull/8168), `2b1f12938` | Gemini 原生/兼容流客户端取消后保留已观察 Usage 和图片输出计数；无输出不伪造账单。保留精简版不同的 Gemini 思考/转换结构，没有回流其他分支依赖。 | 取消/错误区分、断流后 Usage 与图片计数、无 Usage 不扣费 |
| B02 | [#8210](https://github.com/Wei-Shaw/sub2api/pull/8210), `4355236e0` | **精简版安全子集**：Anthropic→Chat/Responses 兼容流取消后继续读上游到最终 Usage，停止客户端写入；180 秒有限排空防无限占用。保留精简版原有 SSE、错误转移、日志及冷却处理。上游额外大规模错误/流泵重构未直接移植，后续仍需单独检查其非断流增强。 | 取消后取得最终 Usage，不能对已断开客户端发送虚假完成事件；无无限排空 |
| B03 | [#8177](https://github.com/Wei-Shaw/sub2api/pull/8177), `9c41b3269` | Antigravity/Gemini 中 prompt 减去 cached 输入 Token 非负裁剪；仅保留精简版可运行测试。 | 负 token、缺失 prompt、cached > prompt |
| B04 | [#8170](https://github.com/Wei-Shaw/sub2api/pull/8170), `1cdecf4a8` | Grok 4.7 Fast 独立目录及 fallback 价；为精简版定价解析器补齐长上下文阈值与包含性。 | 普通/fast 价独立、>=200k Token 档位 |
| B05 | [#7792](https://github.com/Wei-Shaw/sub2api/pull/7792), `b6a157c19` | OpenAI WS v2 passthrough 图片输入、image_gen 使用量在单轮聚合并透传至计费适配器；保留本地已有的轮次状态。 | 图片输入与生成 Token 不被当成纯文本或遗失 |
| B06 | [#7779](https://github.com/Wei-Shaw/sub2api/pull/7779), `73381fc9c` | 各 WS turn 重取合法同组 API Key 的最新 Group 快照，同轮利润检查/账单价格一致；在 lite 无订阅字段，取消上游订阅类型依赖。 | 第二轮调价即时生效、Key 换组/查询失败回退 |
| B07 | [#7573](https://github.com/Wei-Shaw/sub2api/pull/7573), `a2007db70` | 渠道图片单价未单独配置时继承上游价卡；显式设置的图片单价优先。 | 默认价与显式覆盖 |
| B08 | [#7380](https://github.com/Wei-Shaw/sub2api/pull/7380), `0416c9ede`；补充参考上游 `ece1ecccc` | 归一化 Anthropic 兼容流缓存读写及输入 Token 统计，最终 Chat/Responses SSE 用量与计费桶一致；只在上游实际报告 usage 时输出用量事件，不伪造 0 用量。 | 多次 delta 幂等、缓存口径、消息完成/EOF 一致 |

## 约束与待办

1. B02 的上游 PR 大量修改现有 handler/stream/pump，存在行为级和大面积文本冲突，**没有整体移植**。本次只处理精简版直接需要的断流排空与计费，不声称包含上游全部错误策略；后续对上游其余错误事件/中断路径另外评估，不可覆盖 lite 既有实现。
2. 旧总台账的 B07/B08 标为“部分覆盖”；本次仅补充经源码核验的剩余必要计费行为。
3. 本批在 `lite/main` 仍用旧安全工具链的情况下自包含 Go 1.26.10、DOMPurify 3.4.17、follow-redirects 1.16.1；这是独立漏洞扫描要求，不表示对第一批 Git 提交有前置依赖。
4. 未经相同最终 SHA 的 GitHub CI/Security Scan 全通过不得标记为已交付；后续添加 PR 地址和测试运行链接。
5. 回滚：独立批次 PR 可整体 revert；B02 有限排空可单独撤销本批的两条兼容流处理及 `gateway_forward_compat_disconnect.go`，不触碰已合并的独立故障转移/账户冷却策略。

## 测试记录

- Go 1.26.10 受影响四包编译：通过。
- B03、B04、B05、B06、B08 定向测试：通过；新增 B01 和 B02 取消上下文用量回归通过。
- 用量透传 `TestAnthropicChatStreamAuthoritativeUsage` 和 `TestHandleResponsesStreamingResponse_NormalizesTerminalUsage`：通过。
- Go 1.26.10 `go test -tags=unit ./...`：**全通过**，含 `internal/service`、`internal/handler`、`internal/pkg/antigravity`、`openai_ws_v2` 及其他所有 Go 包。
- 前端 pnpm 9.15.9 `install --frozen-lockfile --ignore-scripts`、`lint:check`、`typecheck`、`test:run --maxWorkers=2 --minWorkers=1`、`build`：**均通过**。
- `bash -n deploy/docker-deploy.sh`、`git diff --cached --check`：通过。
- GitHub CI 与 Security Scan：**待最终分支推送并由独立 PR 运行**；不在本地结果中冒称远程验证完成。
