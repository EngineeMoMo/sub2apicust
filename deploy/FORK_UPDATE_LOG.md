# sub2api 定制派生仓更新台账

最后更新：2026-10-24（Asia/Shanghai）。本页记录官方来源、同步基线及本仓发布；Kiro第三方参考的固定SHA与局部移植另见[KIRO_REFERENCE_TRACKING.md](KIRO_REFERENCE_TRACKING.md)。

## 2026-10-24 DOMPurify安全补修（待发布门禁）

用户要求修复Kiro发布版安全阻塞；局部移植官方`8286268047cbbca4e7f5fbb63ff0890c1ea37cf5`，将DOMPurify直接依赖、override下限和锁文件统一为3.4.17，修复GHSA-cg2g-xh64-v4q7（critical）及GHSA-mrv2-3pqr-j5jr（high）。npm官方元数据与上游锁文件SHA512一致；其他依赖未改，例外仍为空。

新增6项安全回归并纳入Makefile关键清单：旧3.4.16复现3项失败，新3.4.17全部通过，普通富文本保留；本地生产依赖审计所有严重度为0。发布代码快照、实际业务SHA的CI/安全/GHCR及摘要待验，不能把本地结果当远程门禁通过。

来源台账[dompurify-20261024.json](../tools/upstream/dompurify-20261024.json)，本机证据`output/dompurify-fix-20261024`。这是局部依赖修复，原官方基线v0.2.17和47提交历史观察不变；未整仓合并v0.2.18，未部署生产或本机，未调用实号，日门槛/261及营销保留未发布。

## 来源与当前基线（此前观察）

| 项目 | 已核实状态 | 取证方式 |
| --- | --- | --- |
| 本仓 origin | https://github.com/EngineeMoMo/sub2apicust ，默认main，公开仓库 | git remote -v及GitHub仓库API |
| 派生关系 | 代码为sub2api定制派生仓；GitHub API的fork=false，没有GitHub parent字段，不能写成已加入官方fork网络 | output/kiro-release-20261024/repository.json |
| 官方 upstream | https://github.com/Wei-Shaw/sub2api ，禁止向该remote推送 | git remote -v |
| 本轮提交前origin/main | cafcd2bac1af98d533b45e19fb3660fb058b940b，与当时本机HEAD相同 | 2026-10-24实际git fetch origin main后rev-parse/rev-list |
| 当前已同步的官方基线 | b0a02a3304a0a78cdd003ddda1dd23ba02c03084，源码VERSION=0.2.17 | git merge-base HEAD upstream/main及git show HEAD:backend/cmd/server/VERSION |
| 本次看到的官方main | 37067e8e340e9440d173a92b4faa05e8ce58368f，源码VERSION=0.2.18 | 2026-10-24实际git fetch upstream main及git show upstream/main:backend/cmd/server/VERSION |
| 尚未同步的官方差异 | 47个提交，含合并/依赖/文档提交，不等于47项独立修复 | git rev-list --count HEAD..upstream/main，发布前记录于output/kiro-release-20261024 |

本轮仅获取远端引用以核实状态，没有合并官方v0.2.18。当前业务仍以v0.2.17为基线；以新的提交SHA核验本仓CI和镜像，不能使用旧本地镜像或其他提交的成功结果。

## 2026-10-24 Kiro接入发布

代码已推送origin/main：首个功能提交3e19f039162515b2a89e0f5cfcea3ed99d4d1dd4，两次静态检查补修后的实际业务提交4b44237acb7a7a1da38b9e0cee5a5f29875d5e48。补修仅调整新增Kiro文件的错误前缀、字符串构建器返回值处理和等价角色switch。下方早期“未提交/未发布”属于历史状态，最新发布状态以本节为准；代码推送不代表生产门禁全部通过。

| 同一业务SHA的检查 | 运行ID | 当前结果 |
| --- | --- | --- |
| CI | [40079131558](https://github.com/EngineeMoMo/sub2apicust/actions/runs/40079131558) | success；完整Linux unit/integration、专用PostgreSQL/Redis、Go lint 0问题、前端71文件1003项、Family产品node 176项、shell及release-helpers通过 |
| Security Scan | [40079131565](https://github.com/EngineeMoMo/sub2apicust/actions/runs/40079131565) | failure；backend-security成功，frontend-security被两项DOMPurify漏洞阻塞 |
| Build custom image (GHCR) | [40079131591](https://github.com/EngineeMoMo/sub2apicust/actions/runs/40079131591) | success；固定SHA镜像、平台/摘要及OCI来源已验真 |

候选镜像ghcr.io/engineemomo/sub2apicust:sha-4b44237，平台linux/amd64，构建VERSION=0.2.17-custom.4b44237ac。index摘要sha256:15fcefd2bd4b3a6f1065782174797566b894a17098a93d0490656a852b74a9eb；平台manifest摘要sha256:9805dc9a8b6212b7005f2ad50cc9f36d73c69d11e72dbe2a5c8dae42ee23bb7d；config摘要sha256:67ebb631e3a00df271d8d757e206f885805268e7e0d0d6cb25369b4960bc0e0d。GHCR原始三层SHA256及OCI revision/source与实际业务SHA匹配，构建日志同时确认前端生产构建、i18n/Family嵌入回归及Linux go build -tags embed成功。证据output/kiro-release-20261024/verified-release.json及business-image-build.log。安全失败仍保留，候选镜像成功不能写成所有发布门禁通过。

提交范围验证来自索引独立快照：前端71文件1003项、类型检查和ESLint（0错误/26警告）通过；后端7包524个定向测试节点通过，独立PostgreSQL CAS含子项10节点通过。静态补修后独立快照再过523节点，并单独补验前后端平台清单一致性1项；没有把缺少前端源码引起的首次跳过当作通过。最后一处仓储错误前缀补修后重新运行独立PostgreSQL CAS，10节点通过、零跳过，测试容器已按ID/名称/标签核实后清理；前端内容与首个功能提交相同。离线参考工具24/24通过；首个功能SHA的远程巡检工作流40077750397仅运行离线测试，远程watch因非人工触发而跳过。

安全阻塞为现有DOMPurify3.4.16：GHSA-cg2g-xh64-v4q7（critical）、GHSA-mrv2-3pqr-j5jr（high）。本轮业务提交前后package.json和pnpm-lock.yaml的Git blob分别相同，未新增审计例外；官方修复828626804已在待同步清单中，本轮尚未迁入。证据output/kiro-release-20261024/business-frontend-security.log（实际最终业务SHA）。生产升级前先处理门禁，当前不提供直接升级命令。

本次范围为原生Kiro OAuth/API Key凭据导入、代理/刷新/CAS/outbox、三协议流式/非流式、后台测试/图标/来源帮助、计量兼容和显式estimated选项，以及参考来源/许可证/人工巡检工具。默认verified不变；不会通过发布代码自动修改实号计量设置。新状态提示的AccountTestModal测试纳入根Makefile关键回归。

日净充值门槛及迁移261、营销素材保留在本地工作树，不纳入本次Kiro业务提交。2026-10-24本地Docker版本0.2.17-custom.kiro-usage-20261024-1092d0243ba0来自含这些并行代码的工作树，不能把它的源码清单当作本次发布提交的完整字节清单。发布前从暂存区导出独立源码快照验证，远程CI和GHCR继续按实际业务SHA核验。

已取得的实号证据仅为testkiro/Sonnet4.5最小生成OK.；billing_ready=false、usage_source=missing，原账号仍verified。estimated的三协议隔离回归已通过，但实号估算收费、精确Token/账单、其他模型权限及图片/工具多轮尚未验收。使用限制见[KIRO_NATIVE.md](KIRO_NATIVE.md)。生产仍由用户自行部署。后续仅文档/来源记录回执使用[skip ci]，不产生新业务镜像；核验对象始终是4b44237acb7a7a1da38b9e0cee5a5f29875d5e48，不把回执SHA的跳过检查当作通过。

## 官方待同步项（按提交主题初筛）

下列是最新差异的审阅入口，尚未逐项移植/验证，不能称本仓已获得这些修复：

- 828626804：DOMPurify 3.4.17及关联安全通告，后续同步优先审阅依赖/安全检查。
- c591412c4、e068c3d02：OpenAI/Codex overload、预输出事件及安全重试。
- 94c7d6b02、4355236e0、2b1f12938：Anthropic兼容流/Gemini客户端断开及不完整上游的计费/收尾；同步时须与Kiro流式用量策略逐项回归。
- e0fa4b94f、47cb84b8b、afbef7c8c、c7bfcd93a：Cloudflare限流判断、Gemini生图开关、DeepSeek图片载荷、工具schema归一化。
- e445cc3e2：按操作及资源范围授权的管理员API Key，需要检查本仓管理员认证/IP/step-up接缝。
- 53002f677、89654873f、67dca9687：调度/模型映射/配额累加性能路径，需要核对本仓平台清单与缓存定制。

完整47个提交SHA/主题及观察状态已持久保存到[官方更新快照](../tools/upstream/sub2api-20261024.json)，全部ported=false；原始取证另存本轮output目录。此清单只表示本次观察到的官方状态，下一次须重新fetch，新增带日期快照并保留历史。

## 下次人工更新流程

1. 用户提出巡检或同步时，读取HANDOFF、CUSTOMIZATIONS、SYNC及本页，确认未提交工作和目标环境；参考项目不设定时任务。
2. fetch origin/main及upstream/main，记录官方HEAD、VERSION、共同祖先、提交范围和有关PR/文件；观察到新提交不等于已经审阅或移植。
3. 若本次任务包含合并，先运行git merge-tree --write-tree HEAD upstream/main演算冲突，再在隔离环境处理；保留定制层、凭据/代理CAS、计费、人工巡检和所有登记接缝。换肤只在theme.css。
4. 改官方文件打[CUSTOM]并登记CUSTOMIZATIONS，固定来源SHA；局部移植逐项记范围，不把整个参考HEAD标为已全部合入。真实配置、凭据、备份及output不进Git。
5. 验证实际将提交的代码；发布后核验同一业务SHA的Linux unit/integration、专用PG/Redis、前端关键回归、lint、安全和embed镜像。记录CI运行ID、固定sha镜像标签、平台/摘要/OCI revision及未验范围。
6. 收工更新本页与HANDOFF四/五、CLAUDE、CUSTOMIZATIONS、SYNC、UPDATE_GUIDE。若随后只做文档回执，注明业务SHA，不能把文档SHA的跳过检查当业务通过；不自动部署生产，不强推。
