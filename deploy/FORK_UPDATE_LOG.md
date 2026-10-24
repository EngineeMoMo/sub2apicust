# sub2api 定制派生仓更新台账

最后更新：2026-10-24（Asia/Shanghai）。本页记录官方来源、同步基线及本仓发布；Kiro第三方参考的固定SHA与局部移植另见[KIRO_REFERENCE_TRACKING.md](KIRO_REFERENCE_TRACKING.md)。

## 来源与当前基线

| 项目 | 已核实状态 | 取证方式 |
| --- | --- | --- |
| 本仓 origin | https://github.com/EngineeMoMo/sub2apicust ，默认main，公开仓库 | git remote -v及GitHub仓库API |
| 派生关系 | 代码为sub2api定制派生仓；GitHub API的fork=false，没有GitHub parent字段，不能写成已加入官方fork网络 | output/kiro-release-20261024/repository.json |
| 官方 upstream | https://github.com/Wei-Shaw/sub2api ，禁止向该remote推送 | git remote -v |
| 本轮提交前origin/main | cafcd2bac1af98d533b45e19fb3660fb058b940b，与本机HEAD相同 | 2026-10-24实际git fetch origin main后rev-parse/rev-list |
| 当前已同步的官方基线 | b0a02a3304a0a78cdd003ddda1dd23ba02c03084，源码VERSION=0.2.17 | git merge-base HEAD upstream/main及git show HEAD:backend/cmd/server/VERSION |
| 本次看到的官方main | 37067e8e340e9440d173a92b4faa05e8ce58368f，源码VERSION=0.2.18 | 2026-10-24实际git fetch upstream main及git show upstream/main:backend/cmd/server/VERSION |
| 尚未同步的官方差异 | 47个提交，含合并/依赖/文档提交，不等于47项独立修复 | git rev-list --count HEAD..upstream/main，发布前记录于output/kiro-release-20261024 |

本轮仅获取远端引用以核实状态，没有合并官方v0.2.18。当前业务仍以v0.2.17为基线；以新的提交SHA核验本仓CI和镜像，不能使用旧本地镜像或其他提交的成功结果。

## 2026-10-24 Kiro接入发布

用户已授权提交、推送本轮代码并维护文档。业务提交SHA及远程检查终态在推送核验后补录；提交前本地验证不等于远程门禁成功。

本次范围为原生Kiro OAuth/API Key凭据导入、代理/刷新/CAS/outbox、三协议流式/非流式、后台测试/图标/来源帮助、计量兼容和显式estimated选项，以及参考来源/许可证/人工巡检工具。默认verified不变；不会通过发布代码自动修改实号计量设置。新状态提示的AccountTestModal测试纳入根Makefile关键回归。

日净充值门槛及迁移261、营销素材保留在本地工作树，不纳入本次Kiro业务提交。2026-10-24本地Docker版本0.2.17-custom.kiro-usage-20261024-1092d0243ba0来自含这些并行代码的工作树，不能把它的源码清单当作本次发布提交的完整字节清单。发布前从暂存区导出独立源码快照验证，远程CI和GHCR继续按实际业务SHA核验。

已取得的实号证据仅为testkiro/Sonnet4.5最小生成OK.；billing_ready=false、usage_source=missing，原账号仍verified。estimated的三协议隔离回归已通过，但实号估算收费、精确Token/账单、其他模型权限及图片/工具多轮尚未验收。使用限制见[KIRO_NATIVE.md](KIRO_NATIVE.md)。生产仍由用户自行部署。

## 官方待同步项（按提交主题初筛）

下列是最新差异的审阅入口，尚未逐项移植/验证，不能称本仓已获得这些修复：

- 828626804：DOMPurify 3.4.17及关联安全通告，后续同步优先审阅依赖/安全检查。
- c591412c4、e068c3d02：OpenAI/Codex overload、预输出事件及安全重试。
- 94c7d6b02、4355236e0、2b1f12938：Anthropic兼容流/Gemini客户端断开及不完整上游的计费/收尾；同步时须与Kiro流式用量策略逐项回归。
- e0fa4b94f、47cb84b8b、afbef7c8c、c7bfcd93a：Cloudflare限流判断、Gemini生图开关、DeepSeek图片载荷、工具schema归一化。
- e445cc3e2：按操作及资源范围授权的管理员API Key，需要检查本仓管理员认证/IP/step-up接缝。
- 53002f677、89654873f、67dca9687：调度/模型映射/配额累加性能路径，需要核对本仓平台清单与缓存定制。

完整提交列表和来源SHA保存在本轮output目录；此清单只表示本次观察到的官方状态，下一次须重新fetch。

## 下次人工更新流程

1. 用户提出巡检或同步时，读取HANDOFF、CUSTOMIZATIONS、SYNC及本页，确认未提交工作和目标环境；参考项目不设定时任务。
2. fetch origin/main及upstream/main，记录官方HEAD、VERSION、共同祖先、提交范围和有关PR/文件；观察到新提交不等于已经审阅或移植。
3. 若本次任务包含合并，先运行git merge-tree --write-tree HEAD upstream/main演算冲突，再在隔离环境处理；保留定制层、凭据/代理CAS、计费、人工巡检和所有登记接缝。换肤只在theme.css。
4. 改官方文件打[CUSTOM]并登记CUSTOMIZATIONS，固定来源SHA；局部移植逐项记范围，不把整个参考HEAD标为已全部合入。真实配置、凭据、备份及output不进Git。
5. 验证实际将提交的代码；发布后核验同一业务SHA的Linux unit/integration、专用PG/Redis、前端关键回归、lint、安全和embed镜像。记录CI运行ID、固定sha镜像标签、平台/摘要/OCI revision及未验范围。
6. 收工更新本页与HANDOFF四/五、CLAUDE、CUSTOMIZATIONS、SYNC、UPDATE_GUIDE。若随后只做文档回执，注明业务SHA，不能把文档SHA的跳过检查当业务通过；不自动部署生产，不强推。
