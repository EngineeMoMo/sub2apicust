# Kiro 接入参考与更新追踪

## 2026-10-24 人工巡检与计量修复（最新）

2026-10-24 14:53（Asia/Shanghai）六仓检查全部成功，均与下表锁定SHA identical、无新提交或Release变化；报告output/kiro-usage-fix-20261024/reference-watch/latest.json。kirocc等在这些既有版本已提供用量回退，不能称本轮发现新的上游修复。局部适配新增[ports-20261024.json](../tools/kiro/ports-20261024.json)，保留完整来源SHA、文件、许可证、差异和测试；sources.lock.json的reviewedCommit/portedCommit未改变。

本轮兼容metering精确字段、metadata优先，连接与计量状态分离；默认verified不变，仅显式estimated才估算并标记来源，不复制缓存模拟/credits美元换算。Go41顶层128含子项、前端152、类型与Linux embed等通过，已更新本机Docker；Sonnet4.5单次实号OK.但未取得精确Token，原verified账号保持，计费与其他模型仍待验收。证据output/kiro-usage-fix-20261024/verification.json及HANDOFF最新状态。未提交推送或生产部署，只人工巡检。


记录日期：2026-10-23，Asia/Shanghai。用户要求按此前建议先审阅 `d-kuro/kirocc` 和 `ZyphrZero/kiro.rs`，并保存参考来源，以便及时采用后续修复。

## 已核实：当前交付范围

本仓已在工作树实现原生Kiro账号、刷新、代理、CAS和三协议/管理入口，并已本机Docker部署；最新计量兼容与实号边界见本页上节和[KIRO_NATIVE.md](KIRO_NATIVE.md)。未提交推送或生产部署；参考项目更新仅按用户要求人工巡检。

六个参考仓库已于 2026-10-23T08:35:03.251Z（北京时间16:35）实际请求 GitHub API，六项成功、零失败，HEAD 均与锁定的观察提交一致；没有发现本次新增提交或发布。原始报告在 `output/kiro-reference-watch/`，后续运行会更新 `latest.*`，带时间的报告保留。

最终24项离线Node测试全通过、零跳过，包含分页/截断、分叉回退、迁移/归档、发布变化、失败游标保护及完整文件落盘的2/0/1状态；另用隔离旧基线实际请求GitHub compare，确认能采集到已知代码差异。工作流YAML可解析，34份文件摘要复核一致。验证范围及命令见 [`verification-20261023.json`](../tools/kiro/verification-20261023.json)；没有借用既有业务CI结果。

## 来源清单与固定版本

机器可读的长期记录：[`sources.lock.json`](../tools/kiro/sources.lock.json)。`reviewedCommit` 仅表示本轮所述核心路径的审阅范围，不能解释为整个第三方项目已审计或可直接用于商用计费。

| 仓库 | 角色与用途 | 观察提交 / 已审阅提交 | 许可证证据 |
| --- | --- | --- | --- |
| [d-kuro/kirocc](https://github.com/d-kuro/kirocc) | 首选 Go 认证、协议、请求与响应转换参考 | `7a2ba670286ca16ae89fe792ea935b3751179421`；核心路径已审阅 | 固定提交 LICENSE 为 Apache-2.0 |
| [ZyphrZero/kiro.rs](https://github.com/ZyphrZero/kiro.rs) | 首选多凭据调度、代理、兼容及用量修复参考 | `4cac2a8b887eacf4c8119558d3c1906e43f8cbd4`；核心路径已审阅 | 固定提交 LICENSE 为 MIT |
| [xiangking/sub2api-kiro](https://github.com/xiangking/sub2api-kiro) | 用户提供的原参考；后台接线、授权和账号导入路径 | `9541e0e9fb64718c15581985c84807232536d39b`；核心路径已审阅 | 固定提交 LICENSE 为 GNU LGPL v3 文本，项目授权范围迁入时复核 |
| [hank9999/kiro.rs](https://github.com/hank9999/kiro.rs) | Rust 原项目的协议、工具图片和模型修复线索 | `22f36e692ad387c46581a84daf0d563347e330a3`；尚未审阅 | GitHub 元数据 MIT |
| [zhujunsan/kiro-gateway](https://github.com/zhujunsan/kiro-gateway) | 参数、思考签名与重试线索 | `e10ced80c9e33128e867ca584456f6c5fd6000bd`；尚未审阅 | GitHub 元数据 AGPL-3.0 |
| [dwgx/KiroStudio](https://github.com/dwgx/KiroStudio) | 管理及多账号维护线索；横幅变化单独分类 | `0c40219cabfc13387d34c5e3a22392e69a2aa56b`；尚未审阅 | GitHub 元数据 NOASSERTION，不能据此认定可复制代码 |

三个已审阅来源的 34 个取证文件，记录了路径、字节数、SHA256 和本地计算的 Git blob SHA1，见 [`review-20261023.json`](../tools/kiro/review-20261023.json)。两个完整源码包来自固定 SHA 的 codeload 归档；归档 SHA256 也已记录。xiangking 前轮另外逐一核对过38个文件的 Git blob SHA1，38/38 与 GitHub 提交树一致。原始源码副本和较大的 API 响应放在忽略的 `output/`，长期基线不依赖这些临时副本存在。

## 已核实：Go 参考的可用思路与迁入接缝

以下文件均相对 kirocc 的固定提交，而非浮动的 main。

| 证据位置 | 源码行为 | 本仓适配要求 |
| --- | --- | --- |
| `internal/auth/refresh.go:24,110,136,159` | 可注入刷新 HTTP 客户端；支持静态 Kiro API Key；从 CLI SQLite 读取 OAuth 凭据，进程内 singleflight 合并刷新 | 读取 CLI 数据库不等于逐请求执行 CLI。接入本仓账号凭据及代理、统一 `OAuthRefreshAPI`；其锁不解决本仓多实例竞争 |
| `internal/auth/refresh.go:159` | 过期时区分 Social / OIDC；刷新结果更新内存缓存，未在该路径写回 SQLite | 本仓需持久化轮换后的 Token，处理保存错误并失效账号及调度缓存，不能直接搬运内存凭据管理器 |
| `internal/kiroclient/client.go:196,318` | 运行端点使用 `runtime.<region>.kiro.dev`；有 API Key TokenType、403刷新、429/5xx退避、200 JSON异常识别和流读取空闲超时 | 认证地区与运行地区分别处理；所有授权、刷新、模型、用量及生成请求接账号代理，并复用本仓取消、超时与换号策略 |
| `internal/kiroproto/frame.go:19`、`eventstream.go:155` | 校验 AWS EventStream 的 prelude/message CRC，解析工具及元数据事件 | 单独覆盖拆帧、截断、CRC损坏、错误帧和取消；不能按普通 SSE 读取上游 |
| `internal/reqconv/build_payload.go:46,173`、`content_text.go:33` | 工具结果与图片有专用结构；响应侧恢复工具名并聚合工具 JSON | 保留工具 ID、历史往返、图片、thinking和非流式工具结果，避免原参考的文本化简化 |

## 已核实：多凭据参考的可用思路与限制

以下文件均相对 ZyphrZero 的固定提交。

| 证据位置 | 源码行为 | 本仓适配要求 |
| --- | --- | --- |
| `src/kiro/token_manager.rs:4501` | 强制刷新先取得进程内刷新锁，再读取最新凭据，并采用该账号的有效代理 | 参考“锁后重读”顺序；本仓继续使用已有按账号协调与数据库存储 |
| `src/kiro/token_manager.rs:2447,4530` | 凭据落盘采用写锁和临时文件rename，但强制刷新调用点对落盘失败仅warn后返回成功 | 不能只看到原子写入函数就认定持久化故障已向上游传播；本仓须验证保存错误及轮换失败后的恢复路径 |
| `src/kiro/provider.rs:1111,1407` | 区分账号级429与普通429；按配置决定冷却换号、原账号重试或直接返回，保留Retry-After | 映射到本仓现有账号冷却和调度，不另建一套独立账号池 |
| `src/kiro/provider.rs:912`、`session_affinity.rs:1` | 上游HTTP接受请求后绑定会话账号；粘性表有容量和TTL，为进程内状态 | 参考失败不绑号的顺序；多副本继续采用本仓粘性及分组隔离约束，HTTP接受也不等于完整流成功 |
| `src/kiro/endpoint/cli.rs:57`、`ide.rs:63` | CLI和IDE路径分别构造请求，地区、profileArn及Token类型有区别 | 先用实际账号确定端点能力，不能把旧CodeWhisperer与新runtime路径混作一个固定接口 |

## 已核实：计量不能直接照搬

1. **两项目都有估算回退。** kirocc的 `internal/respconv/usage.go:21` 按元数据/计量、预先计数、上下文百分比回退，输出可按字符估算；Zyphr的 `src/anthropic/stream.rs:1489,1512` 优先 `metadataEvent.tokenUsage`，缺失时使用请求/上下文估算及本地CacheMeter。`src/anthropic/cache_metering.rs:1` 明确属于本地模拟。代码声明“精确”或测试夹具出现某字段，不能证明实际账号响应一定提供该字段。
2. **kirocc缓存输入口径需改。** `internal/kiroproto/eventstream.go:175` 将 `InputTokens` 设为 `uncachedInputTokens + cacheReadInputTokens`，而 `respconv/usage.go:40` 又单独输出 `cache_read_input_tokens`。例如原始13未缓存、3缓存读、2缓存写、17输出，该路径会把 `input_tokens` 置16。若直接采用本仓Anthropic分项计费，会有缓存读重复进入输入价的风险；迁入应保留13/3/2/17四项并做账单回归。这是源码口径判断，尚未发生本仓Kiro实扣。
3. **credits聚合语义仍需实测。** kirocc `respconv/event_processor.go:83` 保留最近一次有效 `Credits`；Zyphr `anthropic/stream.rs:1672` 将多个 `meteringEvent.usage` 累加，但最终输出还涉及最近一次完整payload。两者行为不同，必须用脱敏原始流和上游账单确定事件是快照还是增量。credit也不能直接当USD。
4. **原参考限制保留在案。** xiangking的 `backend/internal/pkg/kiro/parse.go:26,382` 估算输出且输入取输出一半；`service/kiro_gateway.go:525` 将OpenAI流式请求改为非流式；历史工具有文本化、刷新/额度代理和保存错误接缝。不能整仓合并后就宣称准确计费及完整兼容。

## 更新检查的运行方式

要求Node.js 22及curl。脚本不需要安装npm依赖，Windows使用curl.exe；curl继承HTTP(S)_PROXY等现有网络环境。GitHub认证可走已有的 `GITHUB_TOKEN` 或 `GH_TOKEN` 环境变量，Token经stdin送入curl，不进入命令行参数或巡检报告。私有Token和代理配置不进来源清单。

```powershell
node --test tools/kiro/check-updates.test.mjs
node tools/kiro/check-updates.mjs
```

默认报告目录为 `output/kiro-reference-watch/`：带时间的JSON/Markdown供追溯，`latest.*`供阅读，`state.json`保存上次成功观察的HEAD及发布摘要。`--out <目录>`可进行独立验证；`--manifest <文件>`可使用隔离测试清单。退出0表示检查成功，退出1表示至少一项检查失败；加 `--fail-on-update` 后，退出2专门表示发现需要审阅的新变化。

检查涵盖新提交及公开Release（含预发布和已知Release正文修订）、仓库迁移、默认分支、归档/禁用和历史分叉/回退。提交及Release最多各10页、每页100项；达到上限或触及GitHub compare的300文件限制时明确要求人工核查，不将截断当作无更新。纯README或横幅图片变化仍保存记录，普通情况下不触发功能更新提醒；未知路径、文件改名和不完整差异保守要求审阅。

网络、权限、限流、删除或JSON异常会记为失败；失败来源不推进观察游标，其他来源仍继续。**任何人工运行的检查脚本都不会改 `reviewedCommit`、`portedCommit` 或自动合并业务代码。** 观察游标只是“已经看到”，不是“已经修复”。

## 仅人工巡检（用户2026-10-23最新要求）

- 已通过自动化工具删除本地任务，ID `kiro`，返回 deleteStatus=deleted。此前每日09:00的安排已取消，不再作为待办。
- [独立GitHub工作流](../.github/workflows/custom-kiro-reference-watch.yml)已移除 schedule。远程检查只接受 workflow_dispatch；push/PR仅跑离线工具回归。工作流仍是只读权限、缓存观察游标并上传报告，未提交推送或运行本轮CI。
- 用户以后明确要求巡检时，再运行本页命令，按受影响文件判断与本仓实现的相关性。不会自动创建任何新定时任务、推进审阅/移植基线或自动合并修复。

## 审阅、fork与移植记录规则

发现更新后，先保存仓库、完整SHA、提交/PR/Release链接、受影响路径、修复触发条件和本仓相关性。纯改名、依赖机器人或横幅变化不称成功能修复；尚未迁入的模块标记“后续接入参考”，不能说“本仓漏洞已修复”。

需要保留独立维护分支时再fork，并记录fork地址和来源SHA；单个修复可先固定提交逐项移植。实际迁入必须将下面字段写入 `review-20261023.json` 的 `ports`，或新的对应日期记录，并从本页链接到后续记录：来源仓库、来源完整SHA、来源文件、关联问题/PR、本仓文件、本仓提交、回归命令与结果、保留的许可证/归属声明。最初审阅记录的 ports 为空，是实施前快照；本轮局部适配已另记[ports-20261023.json](../tools/kiro/ports-20261023.json)，包含来源文件到本仓路径及测试映射，两份原许可证保存在tools/kiro/licenses。所有来源的 portedCommit 仍为null，明确未把整个参考HEAD当作已完全迁入。

人工审阅新提交后才能改 `reviewedCommit`，实际移植且测试通过后才能改 `portedCommit`。如果只移植了部分修复，逐项记录，不把整个上游HEAD当作已全部合入；保留尚未迁入的差异。改本仓上游文件时继续按 `[CUSTOM]` 与CUSTOMIZATIONS登记接缝。

## 本轮适配与后续验收

已新增独立协议包、刷新与CAS仓库、三协议原生网关、管理连接测试及前端JSON导入组件；上游接缝逐处登记[CUSTOMIZATIONS](../CUSTOMIZATIONS.md)，验证记录在output/kiro-native-20261023/verification.json。手动更新来源仍以本页前述固定SHA为基线，未在本轮收尾再次运行远程巡检。

2026-10-24已追加metering兼容和显式估算政策，默认仍要求可识别的精确Token；生成、计量及账单验收须分别判断，详细限制以KIRO_NATIVE.md最新说明为准。

## 不知道 / 尚待取得的证据

本机已实际验证OAuth账号Sonnet4.5最小生成，但未取得可识别的精确Token；实号工具多轮/图片、其他模型权限/地区、精确用量与上游账单仍待证据。未运行第三方项目构建/测试，未fork或整仓合并，未提交推送本轮源码、同SHA远程CI或生产部署。离线估算回归及本机健康不能代替真实计费验收。
