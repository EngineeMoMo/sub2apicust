# 分组开放规则审查记录（2026-10-15）

## 用户追加的到期删除语义（本地补修，未发布）

最终本地门禁：真实PG等后端专项137含子项／41顶层／5包零失败零跳过（到期23项）；另handler6项通过。前端53文件821项、其中分组31项及最终类型检查通过，lint0错误／26既有警告，生产构建成功。Linux全量58包11970顶层／23371含子项通过，51个测试因外部服务等条件跳过，embed编译成功；真实PG补充覆盖本功能被跳过的PG场景。完整命令、源码／证据哈希和范围见verification.json，不代表生产现场收费验收或已发布。

原版到期只在PrepareKey内切换有效分组，没有持久迁移／删组，已核对源码确认缺口。现新增后台expiry store／worker及迁移258，在同事务平移Key和显式授权后软删源组，保留历史／余额／Key属性；失败回滚，多实例锁、并发绑定、旧认证快照、到点／启动补处理／通知唤醒、目标已到期链及目标自指兜底都有回归。仅同平台按量组适用，订阅／包号权益不自动合并。失败组不阻挡未来边界；升级首次启动会处理已有到期规则，须备份和审阅。257不改，无真实线上规则变更或部署。

倍率核对进一步沿实际`APIKeyService.GetAvailableGroups → CustomGroupRate → rateAt`路径回归，覆盖时段前1、起点／时段内0.8、终点1且基础倍率不变；比只看管理列表更强，但仍没有test真实收费账单，不能宣称现场收费通过。最终新功能及全量验证以output/group-access-fix-20261015/verification.json为准，下方103／29为追加删除功能前的阶段结果。

## 21:17后补修状态（优先于下方历史审查）

R1–R4已在本地修复并纳入正式回归，尚未提交／发布／部署。R1新增统一fallback服务与handler，目标门槛、授权、到期、活动倍率、RPM及订阅／路由上下文使用目标分组；Messages在客户端识别后切组，其他协议在鉴权装配订阅前切组，ForcePlatform保留，目标模型白名单与渠道映射随后解析。R2拆分通知消费和单个串行Ping，使用原连接deadline与关闭解除阻塞；真实PG消费者延迟／通知积压三轮恢复、断连接、Stop已有正式测试。R3保留同组RPM快照，目标组按用户缓存覆盖值（含nil与0），用户通知失效，100次热请求不再逐次查RPM。R4已有异步图片GET结果保留原鉴权，不执行新增金额门槛／到期切组，POST仍受门槛限制。

证据在output/group-access-fix-20261015：targeted-postgres-final.jsonl含103项（32顶层、5包、零失败／跳过，真实PG18实际执行）；另handler-fallback-final.jsonl含6项模型／上下文回归通过，前端分组29项及类型／定向lint通过；完整验证与源码哈希见verification.json。下方失败复现与行号保留为修复前历史，旧794479760门禁不代表这轮源码已发布。

同时确认独立的管理列表显示缺陷：列表原本只读基础rate_multiplier，没有活动当前值。线上test活动中用户可选分组下拉显示0.8x，21:17:27结束后完整刷新恢复1x；管理列表基础值1x并不能证明计费失败。本地已新增批量只读时段接口和共用服务器时钟显示，沿用原收费逻辑，不定时写基础倍率。没有test真实账单，20:30原规则已被用户在失败后删除，不能把当前关闭状态当成根因。

## 范围与状态

用户要求「仔细核对下有无漏洞」。审查基线为 `be769bf34df72b0c6ae3dc0d36f42254d3be3cec`，相关业务源码与已发布 `794479760593961c08ea842095543ec1d0bd58c5` 相同，因此以下问题也存在于 `sha-7944797` 对应源码。本轮仅审查、复现及更新交接，四项发现尚未修复，未提交推送、构建新镜像或操作生产／本机8080。

审查覆盖规则管理权限与校验、月净充值SQL与退款状态、缓存失效／通知监听、到期路由、分时倍率、用户专属倍率、批量生图、Responses WebSocket及其他网关降级路径。

## 已核实发现

### R1 / P1：既有降级路由可以绕过目标分组门槛

- 位置：`backend/internal/service/gateway_scheduling.go:926`，特别是945处直接返回目标组；初始规则只在 `backend/internal/server/middleware/api_key_auth.go:168` 处理。
- 前提：管理员将可用分组A设置为 Claude Code 专用，并将受充值／余额门槛限制的B配置为非Claude客户端降级组。用户可以使用A，但不符合B的门槛。
- 最小复现：直接将Key指向B时 `PrepareKey` 返回 `GROUP_ACCESS_REQUIREMENTS`；相同用户从A进入非Claude客户端调度，实际查询的组ID为B，成功选出只属于B的账号，未报门槛错误。余额门槛与月充值门槛各一例均复现。测试账号未提供上游密钥，没有发送真实上游请求。
- 原因：`resolveGatewayGroup` 在入口规则校验之后再次切组，没有执行新规则。目标组的到期规则也未在该函数处理。`gateway_handler.go:1032` 的 prompt-too-long 兜底也直接复制目标组后只做计费资格检查；后者属于源码确认的同类遗漏，本轮没有构造完整上游错误链复现。
- 修复方向：让所有后续切组路径统一解析目标有效组及门槛，使用现有缓存；同步更新调度、计费和上下文。既有Claude降级的权限语义须明确，不能只修到期兜底入口，也不能用每请求聚合SQL兜底。

### R2 / P1：同步心跳与通知队列可能互相等待，持续停用资格缓存

- 位置：`backend/internal/service/custom_group_access_listener.go:64`，通知消费循环内同步执行 `listener.Ping()`。
- 真实 PostgreSQL 18 故障注入：测试暂持缓存锁1.1秒模拟通知消费者延迟，并发送1000条不同通知。释放锁、停止发通知4秒后，健康租约仍已过期。
- goroutine堆栈证明互相等待：服务停在 `listener.Ping` → `ExecSimpleQuery` 等回复；pq的两层读取／转发goroutine分别停在 `notify.go:205` 和 `notify.go:842` 向已满通知队列发送。服务正在等待Ping，无法继续消费队列；读取线程因队列已满，无法读到Ping回复。
- 后果：3秒健康租约到期后 `custom_group_access_cache.go:69` 直接回源；该分支不经过singleflight，新的资格请求持续访问数据库。正常停机也可能受阻，因为pq的Ping持有listener锁，Close需要该锁；停机后果由源码推断，未额外执行会挂住的Stop验收。
- 边界：不施加消费者延迟的80万条通知突发测试通过。因此不能声称普通流量必然触发，也未测得生产触发频率。故障注入及互等堆栈已确认问题确实存在。
- 修复方向：通知消费不能同步等待心跳；心跳需单独且限制为一个在途任务，并设计超时／断线关闭和可终止的清理。保留失联回源失败拒绝的原则，补测试队列背压、恢复和Stop。

### R3 / P2：用户快照替换丢失RPM覆盖值，热缓存请求仍逐次查库

- 位置：`backend/internal/service/custom_group_access.go:248`；后续 `billing_cache_service.go:946`。
- 前提：分组启用了门槛、到期或福利规则，用户存在已加载的专属RPM覆盖值。
- 原因：原认证快照包含 `UserGroupRPMOverride`，`PrepareKey` 用普通用户仓库返回的对象替换 `prepared.User`，该对象不包含按分组加载的RPM覆盖值。后续 `checkRPM` 看到nil便调用仓库；实际仓库 `user_group_rate_repo.go:155` 执行SQL，没有额外缓存。
- 复现：原认证快照执行100次RPM检查，仓库查询计数为0；同组、资格热缓存执行100次 `PrepareKey` 加RPM检查，查询计数为100。这里测量的是仓库桩调用数，并通过仓库源码确认对应SQL路径，未把它当作生产数据库吞吐测量。
- 修复方向：同组保留合法快照；切组时按目标组缓存加载RPM覆盖值，不能照搬原组的覆盖值。基准须覆盖完整资格加计费入口，不能只统计规则和月充值聚合次数。

### R4 / P2：消费后跌破门槛，已付费异步图片结果也被拦截

- 位置：`backend/internal/server/middleware/api_key_auth.go:168` 先执行门槛检查，190处才判断 `isAsyncImageTaskRead`。
- 复现：余额门槛100，模拟生成已消费后余额95，携带有效Key访问 `GET /v1/images/tasks/imgtask_owned`，返回403 `GROUP_ACCESS_REQUIREMENTS`，没有进入任务读取handler。
- 影响：例如生成前余额105、生成消费10，用户的结果轮询会失败。实际图片任务查询按用户ID和Key ID校验归属，且现有注释明确要求已完成生成消耗余额后仍可读取结果。
- 修复方向：区分新增付费请求与已拥有任务的只读查询。保留Key、用户、IP和任务归属鉴权，避免新分组门槛使已有结果不可取；不能把整个网关鉴权一并跳过。

## 证据与复现

本机 Go1.27.2 windows/amd64。临时测试使用 `go test -overlay` 注入测试文件，未改生产源码或正式测试目录，也没有把预期失败的用例混入常规测试套件。

证据目录：`output/group-access-review-20261015/`（gitignore内本地证据）。

- `service_review_test.go`、`middleware_review_test.go`、`overlay.json`：可重跑的最小复现。
- `reproduction-final.jsonl`：R1两个分支、R3和R4按正确行为断言均失败；无消费者延迟的80万通知测试通过。
- `listener-backlog.jsonl`、`listener-stacks.txt`：R2真实PG故障注入失败及互等堆栈。
- `existing-tests.jsonl`：不加overlay复跑现有分组专项，含子项76项通过、零失败、零跳过；包括真实PG退款／并发规则和双实例缓存。旧测试通过不覆盖上述新复现。
- `verification.json`：基线、文件摘要与测试计数。
- `container-cleanup.log`：隔离容器清理结果。专用数据库名为 `group_access_test`，随机端口仅监听127.0.0.1，使用tmpfs；完整容器ID及任务标签核对后已删除。

复现命令（在backend目录，先设置指向独立测试库的 `CUSTOM_GROUP_ACCESS_TEST_DSN`；overlay保存的是本工作区绝对路径）：

```powershell
go test -overlay ../output/group-access-review-20261015/overlay.json -tags=unit ./internal/service ./internal/server/middleware -run '^TestCustomGroupAccessReview' -count=1 -timeout=60s -json
go test -tags=unit ./internal/service ./internal/server/middleware -run '^TestCustomGroup(Access|RateSchedule)' -count=1 -timeout=60s -json
```

## 不知道／尚未验证

不知道生产是否配置了触发R1的降级关系，也不知道是否已出现R2队列互等、真实P95变化或费用损失；需要生产配置、请求日志、goroutine／数据库负载等只读证据。此轮没有发现可验证的管理API未授权写入或时段边界计算错误，但不据此保证不存在其他漏洞。退款后异步通知存在传播窗口，仍不承诺事务提交后的下一条并发请求绝对零延迟。
