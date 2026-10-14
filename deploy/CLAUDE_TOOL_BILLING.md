# Claude 付费工具计费（2026-10-14）

本轮支持 Claude 内置网页搜索 `web_search` 的附加费。搜索产生的输入/输出/缓存 Token 继续按原有模型价卡和阶梯计费，工具费用另外叠加。普通客户端 `tool_use`、工具声明以及搜索启动事件都不会单独触发收费。

## 管理员如何设置

后台 → 分组管理 → 新建/编辑 Anthropic（Claude）或 composite 分组 → **Claude 网页搜索工具附加费**。

- 单价单位是 USD/次的基础价格，复用分组字段 `web_search_price_per_call`。
- 留空使用 $0.01/次；填写 0 表示免费；可以填写上游对应的基础搜索单价或自定售价。
- 实扣工具费 = 实际搜索次数 × 基础单价 × 有效倍率。
- 有效倍率沿用该请求 Token 的口径：用户专属倍率优先于分组倍率，再叠加高峰倍率。账号倍率不决定客户售价。
- OpenAI 分组现有 Codex 网页搜索表单及按次请求行为保留；如果使用 OpenAI 网关转发 Claude 原生 Messages，Claude 搜索附加费复用该分组的搜索单价并叠加到 Token 费用。
- composite 分组只配置一份该单价；路由到 Claude 原生协议且收到实际搜索计量时生效。

例如一次成功搜索，基础单价 0.01，有效倍率 0.15，工具附加费为 0.0015。原本 Token 实扣 0.0008559 的请求变为 0.0023559。该示例只验证收费公式，不是读取用户上游账单所得的真实成本。

## 次数证据与边界

1. 优先读取上游原始 JSON/SSE 的 `usage.server_tool_use.web_search_requests`。只接受非负整数，不接受字符串、负数、分数或溢出值。
2. 流式开始和结束事件通常重复或累计报告次数，取最大累计值，不按事件相加；明确返回 0 时保留权威 0。
3. 没有有效计数时，只从实际 `web_search_tool_result` 且 `content` 是成功结果数组的内容块回退计数；按 `tool_use_id` 去重。错误结果对象不产生工具费，搜索启动/普通客户端工具不产生工具费。
4. 上游同时不返回有效次数和成功结果时，无法核实搜索量，不猜测收费。本轮未给其他工具（网页抓取、代码执行等）添加无依据的固定费。
5. 已覆盖 Claude 通用网关、Anthropic API Key 直通的 JSON/SSE、Claude→Responses 的缓冲/流式转换，以及 OpenAI 网关的 Claude 原生 Messages 转发。上游若先转换为另一套协议并删除全部 Claude 工具计量，需要另取原始响应再补适配。

## 账单与实际扣除

客户总费用和实扣金额都包含工具费。余额、订阅消费、Key 额度和速率额度复用既有结算及幂等链路。账号统计规则先计算模型费，再补回工具基础费，最后按账号倍率显示统计成本；该数字仍为本站估算，不是读取上游实扣。

管理员和用户使用日志的费用提示展示工具次数、基础单价、原始工具费、实扣工具费，并注明已含在本次总额。管理员 XLSX 和用户 CSV 均新增工具用量/费用 JSON 列。快照示例：

```json
{"web_search":{"calls":1,"unit_price":0.01,"total_cost":0.01,"actual_cost":0.0015}}
```

包号免扣和扣费失败的日志保留工具次数及参考价格，工具实扣显示 0。工具费不独立再扣一遍。

## 升级与维护

新增迁移 `256_custom_claude_tool_usage.sql`，为 `usage_logs` 新增 nullable JSONB `tool_usage`。应用启动时走正常迁移；旧日志保持 NULL，历史费用不补扣。前后端应同版本发布，生产升级前按现有手册备份数据库。本轮实现不代表生产已更新。

主要新增实现为 `backend/internal/service/custom_claude_tool_billing.go`、`backend/internal/repository/custom_tool_usage_json.go`、`frontend/src/custom/components/UsageToolCosts.vue`。全部上游接缝登记在 CUSTOMIZATIONS.md；Ent 从 schema 生成，不手改生成文件。

测试包括默认/自定义/免费价格、有效倍率、SSE 累计去重、失败工具免费、普通工具不收费、双原生网关叠加、账号统计、余额/订阅/Key 额度、JSON/DTO/前端表单保存与展示。隔离 PostgreSQL 验证迁移重跑、同步/批量/最佳努力写入、旧日志兼容及余额/订阅幂等。

## 取证来源

Anthropic 官方 SDK： https://github.com/anthropics/anthropic-sdk-go/blob/main/message.go （ServerToolUsage.WebSearchRequests、Usage.ServerToolUse）。

Anthropic 官方工具文档： https://platform.claude.com/docs/en/agents-and-tools/tool-use/web-search-tool （每千次成功搜索 $10，另计 Token；搜索错误不收工具费）。2026-10-14 下载快照保存在被 Git 忽略的 `output/newapi-ledger-20261014/anthropic-*`。

真实用户上游工具种类、基础单价、有效倍率和实际响应仍需现场确认；该功能不自动同步 New API 价卡，也不能保证不同充值价值、阶梯和其他附加费造成的全部差价都消失。
